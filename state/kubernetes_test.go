package state

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/api/core/v1"

	"github.com/camptocamp/terraboard/config"
)

func TestNewKubernetes(t *testing.T) {
	k8sConfig := config.KubernetesConfig{
		Namespace:    "test-namespace",
		SecretSuffix: "tfstate",
	}

	// This will fail in unit tests since we don't have a real cluster
	// but we're testing the configuration parsing
	_, err := NewKubernetes(k8sConfig, false)
	if err == nil {
		t.Error("Expected error when no valid kubeconfig is available")
	}
}

func TestNewKubernetesEmpty(t *testing.T) {
	k8sConfig := config.KubernetesConfig{}

	provider, err := NewKubernetes(k8sConfig, false)
	if err != nil {
		t.Errorf("Expected no error for empty config, got: %v", err)
	}
	if provider != nil {
		t.Error("Expected nil provider for empty config")
	}
}

func TestNewKubernetesCollection(t *testing.T) {
	c := &config.Config{
		Kubernetes: []config.KubernetesConfig{
			{}, // Empty config should be skipped
		},
	}

	providers, err := NewKubernetesCollection(c)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(providers) != 0 {
		t.Errorf("Expected 0 providers, got %d", len(providers))
	}
}

// Mock tests using fake Kubernetes client
func createFakeKubernetesProvider() *Kubernetes {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	return &Kubernetes{
		client:    clientset,
		namespace: "default",
		suffix:    "tfstate",
		labels:    map[string]string{},
		noLocks:   false,
	}
}

func TestGetStatesEmpty(t *testing.T) {
	k8s := createFakeKubernetesProvider()

	states, err := k8s.GetStates()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("Expected 0 states, got %d", len(states))
	}
}

func TestGetStatesWithSecrets(t *testing.T) {
	k8s := createFakeKubernetesProvider()

	// Create test secrets
	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-state-tfstate",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"tfstate": []byte(`{"version": 4, "terraform_version": "1.0.0"}`),
		},
	}

	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "another-state-tfstate",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"tfstate": []byte(`{"version": 4, "terraform_version": "1.0.0"}`),
		},
	}

	// Add secrets to fake client
	k8s.client.CoreV1().Secrets("default").Create(nil, secret1, metav1.CreateOptions{})
	k8s.client.CoreV1().Secrets("default").Create(nil, secret2, metav1.CreateOptions{})

	states, err := k8s.GetStates()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("Expected 2 states, got %d", len(states))
	}

	expectedStates := []string{"another-state", "test-state"}
	for i, state := range states {
		if state != expectedStates[i] {
			t.Errorf("Expected state %s, got %s", expectedStates[i], state)
		}
	}
}

func TestGetKubernetesVersions(t *testing.T) {
	k8s := createFakeKubernetesProvider()

	// Create test secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-state-tfstate",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Data: map[string][]byte{
			"tfstate": []byte(`{"version": 4, "terraform_version": "1.0.0"}`),
		},
	}

	k8s.client.CoreV1().Secrets("default").Create(nil, secret, metav1.CreateOptions{})

	versions, err := k8s.GetVersions("test-state")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}
	if versions[0].ID != "current" {
		t.Errorf("Expected version ID 'current', got %s", versions[0].ID)
	}
}

func TestGetKubernetesLocksEmpty(t *testing.T) {
	k8s := createFakeKubernetesProvider()

	locks, err := k8s.GetLocks()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(locks) != 0 {
		t.Errorf("Expected 0 locks, got %d", len(locks))
	}
}

func TestGetKubernetesLocksDisabled(t *testing.T) {
	k8s := createFakeKubernetesProvider()
	k8s.noLocks = true

	locks, err := k8s.GetLocks()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(locks) != 0 {
		t.Errorf("Expected 0 locks when disabled, got %d", len(locks))
	}
}

func TestGetKubernetesLocksWithAnnotations(t *testing.T) {
	k8s := createFakeKubernetesProvider()

	// Create test secret with lock annotations
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-state-tfstate",
			Namespace: "default",
			Annotations: map[string]string{
				"terraform.io/lock-id":        "test-lock-id",
				"terraform.io/lock-operation": "apply",
				"terraform.io/lock-info":      "Terraform apply",
				"terraform.io/lock-who":       "user@example.com",
				"terraform.io/lock-version":   "1.0.0",
				"terraform.io/lock-created":   time.Now().Format(time.RFC3339),
			},
		},
		Data: map[string][]byte{
			"tfstate": []byte(`{"version": 4, "terraform_version": "1.0.0"}`),
		},
	}

	k8s.client.CoreV1().Secrets("default").Create(nil, secret, metav1.CreateOptions{})

	locks, err := k8s.GetLocks()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(locks) != 1 {
		t.Errorf("Expected 1 lock, got %d", len(locks))
	}

	lock, exists := locks["test-state"]
	if !exists {
		t.Error("Expected lock for 'test-state' not found")
	}
	if lock.ID != "test-lock-id" {
		t.Errorf("Expected lock ID 'test-lock-id', got %s", lock.ID)
	}
	if lock.Operation != "apply" {
		t.Errorf("Expected operation 'apply', got %s", lock.Operation)
	}
}