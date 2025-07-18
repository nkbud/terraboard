# OAuth2 Proxy Integration with Terraboard

This document provides comprehensive guidance for integrating OAuth2 Proxy as an authentication gateway for Terraboard with Okta as the backend Identity Provider (IdP).

## Overview

OAuth2 Proxy is a reverse proxy that provides authentication using various providers including Okta. This integration ensures all traffic to the Terraboard application is authenticated while maintaining the existing user experience.

## Prerequisites

- Kubernetes cluster with Ingress controller (nginx recommended)
- Okta tenant with admin access
- Helm 3.x installed
- DNS configuration for your Terraboard domain

## Okta Configuration

### 1. Create Okta Application

1. Log in to your Okta Admin Console
2. Navigate to **Applications** > **Applications**
3. Click **Create App Integration**
4. Choose **OIDC - OpenID Connect**
5. Select **Web Application**
6. Configure the application:
   - **App integration name**: `Terraboard`
   - **Grant type**: Authorization Code
   - **Sign-in redirect URIs**: `https://your-terraboard-domain.com/oauth2/callback`
   - **Sign-out redirect URIs**: `https://your-terraboard-domain.com/oauth2/sign_out`
   - **Controlled access**: Choose based on your organization's requirements

### 2. Retrieve Okta Configuration

After creating the application, note the following values:
- **Client ID**: Found in the application's General tab
- **Client Secret**: Found in the application's General tab
- **Issuer URL**: Your Okta domain's issuer URL (e.g., `https://your-domain.okta.com/oauth2/default`)

### 3. Configure User Assignment

1. Navigate to the **Assignments** tab of your Terraboard application
2. Assign users or groups who should have access to Terraboard
3. Optionally configure custom claims if needed

## Helm Chart Configuration

### 1. Generate Cookie Secret

Generate a secure cookie secret for OAuth2 Proxy:

```bash
openssl rand -base64 32
```

### 2. Create Values File

Create a `values-oauth2.yaml` file with your configuration:

```yaml
# OAuth2 Proxy configuration
oauth2Proxy:
  enabled: true
  image:
    repository: quay.io/oauth2-proxy/oauth2-proxy
    tag: v7.6.0
    pullPolicy: IfNotPresent
  
  config:
    # Okta OIDC configuration
    provider: "oidc"
    oidcIssuerUrl: "https://your-domain.okta.com/oauth2/default"
    clientId: "your-client-id"
    clientSecret: "your-client-secret"
    
    # Cookie configuration
    cookieSecret: "your-generated-cookie-secret"
    cookieDomain: "your-terraboard-domain.com"
    cookieSecure: true
    cookieName: "_oauth2_proxy"
    cookieExpire: "168h"  # 7 days
    
    # Email domain restrictions (optional)
    emailDomains: "*"  # Or restrict to specific domains like "yourcompany.com"
    
    # Redirect URL
    redirectUrl: "https://your-terraboard-domain.com/oauth2/callback"
    
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
      - "--whitelist-domain=your-terraboard-domain.com"
  
  # Resource configuration
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi

# Ingress configuration for OAuth2 Proxy
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod  # If using cert-manager
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: your-terraboard-domain.com
      mainPaths: ['/']
      swaggerPaths: ['/docs']
  tls:
    - secretName: terraboard-tls
      hosts:
        - your-terraboard-domain.com

# Standard Terraboard configuration
terraboard:
  base_url: "/"

# Database configuration
db:
  host: "postgresql"
  name: "terraboard"
  user: "terraboard"
  password: "your-db-password"
  sslmode: "require"

# AWS configuration (if using S3 backend)
aws:
  region: "us-east-1"
  bucket: "your-terraform-state-bucket"
  dynamodb_table: "terraform-state-lock"
  file_extension: ".tfstate"
```

### 3. Deploy with Helm

```bash
# Add the repository (if not already added)
helm repo add terraboard https://nkbud.github.io/terraboard

# Install or upgrade with OAuth2 Proxy
helm upgrade --install terraboard terraboard/terraboard \
  -f values-oauth2.yaml \
  --namespace terraboard \
  --create-namespace
```

## Security Considerations

### 1. Cookie Secret Security

- Use a strong, randomly generated cookie secret
- Store the cookie secret securely (consider using external secret management)
- Rotate the cookie secret periodically

### 2. TLS Configuration

- Always use HTTPS in production
- Use valid SSL certificates (Let's Encrypt recommended)
- Enable secure cookies (`cookieSecure: true`)

### 3. Domain Restrictions

- Configure `cookieDomain` to match your Terraboard domain
- Use `emailDomains` to restrict access to specific email domains if needed
- Consider using `whitelist-domain` for additional security

### 4. Network Security

- Use network policies to restrict traffic between pods
- Consider using service mesh for additional security layers
- Implement proper RBAC for the OAuth2 Proxy service account

## Advanced Configuration

### 1. Custom Claims and Headers

If you need custom claims from Okta, you can configure OAuth2 Proxy to pass additional headers:

```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--set-xauthrequest"
      - "--pass-user-headers"
      - "--pass-authorization-header"
      - "--set-authorization-header"
```

### 2. Skip Authentication for Specific Paths

To skip authentication for health checks or specific endpoints:

```yaml
oauth2Proxy:
  config:
    skipAuthRegex:
      - "^/health$"
      - "^/metrics$"
```

### 3. Multiple Ingress Controllers

If using multiple ingress controllers, specify the correct class:

```yaml
ingress:
  annotations:
    kubernetes.io/ingress.class: "nginx"  # or "traefik", etc.
```

## Troubleshooting

### 1. Common Issues

**Authentication Loop**: Check that the redirect URL in Okta matches the one configured in OAuth2 Proxy.

**Cookie Issues**: Ensure the cookie domain matches your Terraboard domain and that cookies are secure in production.

**OIDC Discovery**: Verify the OIDC issuer URL is correct and accessible from your cluster.

### 2. Debug Logging

Enable debug logging to troubleshoot issues:

```yaml
oauth2Proxy:
  config:
    logLevel: "debug"
    requestLogging: true
    authLogging: true
```

### 3. Health Checks

OAuth2 Proxy provides health check endpoints:
- `/ping` - Basic health check
- `/ready` - Readiness check

## Monitoring and Metrics

OAuth2 Proxy provides metrics that can be scraped by Prometheus:

```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--metrics-address=0.0.0.0:44180"
```

Add a service monitor for Prometheus:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: oauth2-proxy-metrics
  labels:
    app.kubernetes.io/name: oauth2-proxy-metrics
spec:
  ports:
  - name: metrics
    port: 44180
    targetPort: 44180
  selector:
    app.kubernetes.io/component: oauth2-proxy
```

## Testing the Integration

### 1. Basic Authentication Flow

1. Navigate to your Terraboard URL
2. You should be redirected to Okta for authentication
3. After successful authentication, you should be redirected back to Terraboard
4. Verify that user information is displayed correctly in Terraboard

### 2. Header Verification

Check that the correct headers are being passed to Terraboard:

```bash
# View OAuth2 Proxy logs
kubectl logs -n terraboard deployment/terraboard-oauth2-proxy

# Check Terraboard logs for user headers
kubectl logs -n terraboard deployment/terraboard
```

### 3. Session Management

Test session persistence and logout functionality:

1. Close browser and reopen - should remain authenticated
2. Test logout URL if configured
3. Verify session expiration works correctly

## Production Deployment Checklist

- [ ] Okta application configured with correct redirect URIs
- [ ] Strong cookie secret generated and stored securely
- [ ] TLS certificates configured and valid
- [ ] DNS configured to point to your ingress
- [ ] Resource limits configured appropriately
- [ ] Monitoring and logging configured
- [ ] Backup and recovery procedures in place
- [ ] Security policies reviewed and implemented
- [ ] Load testing performed
- [ ] Documentation updated for your specific environment

## References

- [OAuth2 Proxy Documentation](https://oauth2-proxy.github.io/oauth2-proxy/)
- [Okta OIDC Documentation](https://developer.okta.com/docs/reference/api/oidc/)
- [Kubernetes Ingress NGINX](https://kubernetes.github.io/ingress-nginx/)
- [Terraboard Documentation](https://github.com/nkbud/terraboard)

## Support

For issues specific to this integration, please check:
1. OAuth2 Proxy logs for authentication issues
2. Ingress controller logs for routing issues
3. Terraboard logs for application issues
4. Okta system logs for IdP-related issues

For additional support, please refer to the respective project documentation or create an issue in the Terraboard repository.