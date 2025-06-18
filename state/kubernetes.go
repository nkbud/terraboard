package state

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/camptocamp/terraboard/config"
	"github.com/camptocamp/terraboard/internal/terraform/states/statefile"
	log "github.com/sirupsen/logrus"
)

// Kubernetes is a state provider type, leveraging Kubernetes Secrets
type Kubernetes struct {
	client    kubernetes.Interface
	namespace string
	suffix    string
	labels    map[string]string
	noLocks   bool
}

// NewKubernetes creates a Kubernetes object
func NewKubernetes(k8sConfig config.KubernetesConfig, noLocks bool) (*Kubernetes, error) {
	if k8sConfig.Namespace == "" && k8sConfig.SecretSuffix == "" {
		return nil, nil
	}

	var config *rest.Config
	var err error

	if k8sConfig.InClusterConfig {
		log.Info("Using in-cluster Kubernetes configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.WithFields(log.Fields{
			"config_path": k8sConfig.ConfigPath,
			"context":     k8sConfig.ConfigContext,
		}).Info("Using kubeconfig for Kubernetes configuration")

		kubeconfig := k8sConfig.ConfigPath
		if kubeconfig == "" {
			kubeconfig = filepath.Join(clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename())
		}

		configOverrides := &clientcmd.ConfigOverrides{}
		if k8sConfig.ConfigContext != "" {
			configOverrides.CurrentContext = k8sConfig.ConfigContext
		}
		if k8sConfig.ConfigContextAuthInfo != "" {
			configOverrides.Context.AuthInfo = k8sConfig.ConfigContextAuthInfo
		}
		if k8sConfig.ConfigContextCluster != "" {
			configOverrides.Context.Cluster = k8sConfig.ConfigContextCluster
		}

		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			configOverrides,
		).ClientConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &Kubernetes{
		client:    clientset,
		namespace: k8sConfig.Namespace,
		suffix:    k8sConfig.SecretSuffix,
		labels:    k8sConfig.Labels,
		noLocks:   noLocks,
	}, nil
}

// NewKubernetesCollection returns a collection of Kubernetes providers
func NewKubernetesCollection(c *config.Config) ([]*Kubernetes, error) {
	var providers []*Kubernetes

	for _, k8sConfig := range c.Kubernetes {
		provider, err := NewKubernetes(k8sConfig, c.Provider.NoLocks)
		if err != nil {
			return nil, err
		}
		if provider != nil {
			providers = append(providers, provider)
		}
	}

	return providers, nil
}

// GetStates returns a list of all Terraform state names stored in Kubernetes Secrets
func (k *Kubernetes) GetStates() ([]string, error) {
	ctx := context.Background()
	
	labelSelector := metav1.LabelSelector{}
	if k.labels != nil && len(k.labels) > 0 {
		labelSelector.MatchLabels = k.labels
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	secrets, err := k.client.CoreV1().Secrets(k.namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var states []string
	for _, secret := range secrets.Items {
		// Check if this secret contains Terraform state
		if _, ok := secret.Data["tfstate"]; ok {
			// Remove the suffix to get the state name
			stateName := strings.TrimSuffix(secret.Name, "-"+k.suffix)
			if stateName != secret.Name { // Only add if suffix was actually removed
				states = append(states, stateName)
			}
		}
	}

	sort.Strings(states)
	return states, nil
}

// GetState retrieves a specific Terraform state from a Kubernetes Secret
func (k *Kubernetes) GetState(stateName, version string) (*statefile.File, error) {
	ctx := context.Background()
	
	secretName := stateName + "-" + k.suffix
	if version != "" {
		// For versioning, we could use annotations or different naming schemes
		// For now, we'll ignore version parameter as Kubernetes doesn't have native versioning
		log.WithFields(log.Fields{
			"state":   stateName,
			"version": version,
		}).Warn("Version parameter ignored for Kubernetes backend - no native versioning support")
	}

	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	stateData, ok := secret.Data["tfstate"]
	if !ok {
		return nil, fmt.Errorf("secret %s does not contain tfstate data", secretName)
	}

	// The state data might be base64 encoded
	var stateJSON []byte
	if decoded, err := base64.StdEncoding.DecodeString(string(stateData)); err == nil {
		stateJSON = decoded
	} else {
		// If base64 decode fails, assume it's already JSON
		stateJSON = stateData
	}

	file, err := statefile.Read(strings.NewReader(string(stateJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Terraform state: %w", err)
	}

	return file, nil
}

// GetVersions returns version history for a state (limited support in Kubernetes)
func (k *Kubernetes) GetVersions(stateName string) ([]Version, error) {
	ctx := context.Background()
	
	secretName := stateName + "-" + k.suffix
	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	// Kubernetes doesn't have native versioning, so we return a single version
	// based on the secret's last modified time
	versions := []Version{
		{
			ID:           "current",
			LastModified: secret.ObjectMeta.CreationTimestamp.Time,
		},
	}

	// Check if there's a last-modified annotation
	if lastModified, ok := secret.Annotations["terraform.io/last-modified"]; ok {
		if t, err := time.Parse(time.RFC3339, lastModified); err == nil {
			versions[0].LastModified = t
		}
	}

	return versions, nil
}

// GetLocks returns lock information from Kubernetes Secret annotations
func (k *Kubernetes) GetLocks() (map[string]LockInfo, error) {
	if k.noLocks {
		return map[string]LockInfo{}, nil
	}

	ctx := context.Background()
	locks := make(map[string]LockInfo)

	labelSelector := metav1.LabelSelector{}
	if k.labels != nil && len(k.labels) > 0 {
		labelSelector.MatchLabels = k.labels
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	secrets, err := k.client.CoreV1().Secrets(k.namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets for locks: %w", err)
	}

	for _, secret := range secrets.Items {
		// Check if this secret has lock annotations
		if lockID, ok := secret.Annotations["terraform.io/lock-id"]; ok {
			stateName := strings.TrimSuffix(secret.Name, "-"+k.suffix)
			if stateName == secret.Name {
				continue // Skip if suffix wasn't found
			}

			lockInfo := LockInfo{
				ID:   lockID,
				Path: stateName,
			}

			if operation, ok := secret.Annotations["terraform.io/lock-operation"]; ok {
				lockInfo.Operation = operation
			}

			if info, ok := secret.Annotations["terraform.io/lock-info"]; ok {
				lockInfo.Info = info
			}

			if who, ok := secret.Annotations["terraform.io/lock-who"]; ok {
				lockInfo.Who = who
			}

			if version, ok := secret.Annotations["terraform.io/lock-version"]; ok {
				lockInfo.Version = version
			}

			if created, ok := secret.Annotations["terraform.io/lock-created"]; ok {
				if t, err := time.Parse(time.RFC3339, created); err == nil {
					lockInfo.Created = &t
				}
			}

			locks[stateName] = lockInfo
		}
	}

	return locks, nil
}