#!/bin/bash
# Quick Start Script for OAuth2 Proxy with Terraboard and Okta
# 
# This script helps you quickly set up OAuth2 Proxy with Terraboard using Okta as the IdP

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    printf "${1}%s${NC}\n" "$2"
}

print_color $BLUE "=== Terraboard OAuth2 Proxy Setup ==="
echo

# Check prerequisites
print_color $YELLOW "Checking prerequisites..."

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    print_color $RED "ERROR: Helm is not installed. Please install Helm 3.x first."
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    print_color $RED "ERROR: kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if openssl is installed
if ! command -v openssl &> /dev/null; then
    print_color $RED "ERROR: openssl is not installed. Please install openssl first."
    exit 1
fi

print_color $GREEN "✓ Prerequisites check passed"
echo

# Gather configuration information
print_color $YELLOW "Please provide the following information:"
echo

read -p "Terraboard domain (e.g., terraboard.example.com): " DOMAIN
if [[ -z "$DOMAIN" ]]; then
    print_color $RED "ERROR: Domain is required"
    exit 1
fi

read -p "Okta domain (e.g., your-company.okta.com): " OKTA_DOMAIN
if [[ -z "$OKTA_DOMAIN" ]]; then
    print_color $RED "ERROR: Okta domain is required"
    exit 1
fi

read -p "Okta Client ID: " CLIENT_ID
if [[ -z "$CLIENT_ID" ]]; then
    print_color $RED "ERROR: Client ID is required"
    exit 1
fi

read -s -p "Okta Client Secret: " CLIENT_SECRET
echo
if [[ -z "$CLIENT_SECRET" ]]; then
    print_color $RED "ERROR: Client Secret is required"
    exit 1
fi

read -p "Namespace (default: terraboard): " NAMESPACE
NAMESPACE=${NAMESPACE:-terraboard}

read -p "Release name (default: terraboard): " RELEASE_NAME
RELEASE_NAME=${RELEASE_NAME:-terraboard}

echo

# Generate cookie secret
print_color $YELLOW "Generating cookie secret..."
COOKIE_SECRET=$(openssl rand -base64 32)
print_color $GREEN "✓ Cookie secret generated"
echo

# Create values file
VALUES_FILE="terraboard-oauth2-values.yaml"
print_color $YELLOW "Creating Helm values file: $VALUES_FILE"

cat > "$VALUES_FILE" << EOF
# OAuth2 Proxy configuration for Terraboard with Okta
oauth2Proxy:
  enabled: true
  image:
    repository: quay.io/oauth2-proxy/oauth2-proxy
    tag: v7.6.0
    pullPolicy: IfNotPresent
  
  config:
    # Okta OIDC configuration
    provider: "oidc"
    oidcIssuerUrl: "https://${OKTA_DOMAIN}/oauth2/default"
    clientId: "${CLIENT_ID}"
    clientSecret: "${CLIENT_SECRET}"
    
    # Cookie configuration
    cookieSecret: "${COOKIE_SECRET}"
    cookieDomain: "${DOMAIN}"
    cookieSecure: true
    cookieName: "_oauth2_proxy"
    cookieExpire: "168h"
    
    # Email domain restrictions
    emailDomains: "*"
    
    # Redirect URL
    redirectUrl: "https://${DOMAIN}/oauth2/callback"
    
    # Upstream service
    upstreams: "http://terraboard:80"
    
    # Security settings
    skipProviderCaCert: false
    setAuthorizationHeader: true
    setXauthrequest: true
    
    # Logging configuration
    requestLogging: true
    standardLogging: true
    authLogging: true
    logLevel: "info"
    
    # Additional OAuth2 Proxy arguments
    extraArgs:
      - "--scope=openid profile email"
      - "--whitelist-domain=${DOMAIN}"
  
  # Resource configuration
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi

# Ingress configuration
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: ${DOMAIN}
      mainPaths: ['/']
      swaggerPaths: ['/docs']
  tls:
    - secretName: terraboard-tls
      hosts:
        - ${DOMAIN}

# Standard Terraboard configuration
terraboard:
  base_url: "/"

# Additional environment variables
additionalEnv:
  - name: TERRABOARD_LOGOUT_URL
    value: "https://${DOMAIN}/oauth2/sign_out"

# Resource configuration for main Terraboard container
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

# Security context
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL

podSecurityContext:
  fsGroup: 1000
  runAsNonRoot: true
  runAsUser: 1000
EOF

print_color $GREEN "✓ Values file created: $VALUES_FILE"
echo

# Display next steps
print_color $BLUE "=== Next Steps ==="
echo

print_color $YELLOW "1. Review and update the generated values file if needed:"
echo "   $VALUES_FILE"
echo

print_color $YELLOW "2. Make sure your Okta application is configured with these URLs:"
echo "   - Sign-in redirect URI: https://${DOMAIN}/oauth2/callback"
echo "   - Sign-out redirect URI: https://${DOMAIN}/oauth2/sign_out"
echo

print_color $YELLOW "3. Create the namespace (if it doesn't exist):"
echo "   kubectl create namespace ${NAMESPACE}"
echo

print_color $YELLOW "4. Deploy Terraboard with OAuth2 Proxy:"
echo "   helm upgrade --install ${RELEASE_NAME} . -f ${VALUES_FILE} -n ${NAMESPACE}"
echo

print_color $YELLOW "5. Wait for deployment to be ready:"
echo "   kubectl wait --for=condition=available deployment/${RELEASE_NAME} -n ${NAMESPACE}"
echo "   kubectl wait --for=condition=available deployment/${RELEASE_NAME}-oauth2-proxy -n ${NAMESPACE}"
echo

print_color $YELLOW "6. Check the status:"
echo "   kubectl get pods -n ${NAMESPACE}"
echo "   kubectl get ingress -n ${NAMESPACE}"
echo

print_color $YELLOW "7. Access Terraboard at: https://${DOMAIN}"
echo

print_color $GREEN "Setup complete! Review the steps above and execute them to deploy Terraboard with OAuth2 Proxy."

print_color $RED "⚠️  Security Notes:"
echo "   - The generated values file contains sensitive information"
echo "   - Consider using external secret management in production"
echo "   - Review all security settings before production deployment"
echo "   - Ensure proper DNS configuration for your domain"