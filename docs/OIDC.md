# OIDC Authentication with Dex

This document explains how to configure Terraboard to use OIDC authentication with Dex.

## Overview

Terraboard now supports OIDC (OpenID Connect) authentication using Kubernetes Dex as an identity provider. This enables secure authentication without requiring external proxy services like oauth2_proxy.

## Features

- **Login Screen**: Vue.js frontend displays a login screen for unauthenticated users
- **OIDC Flow**: Complete authorization code flow with state validation
- **Token Storage**: Secure cookie-based token storage
- **Logout**: Proper logout functionality that clears authentication cookies
- **Backward Compatibility**: Maintains support for existing proxy-based authentication

## Configuration

### Terraboard Configuration

Add the following OIDC configuration to your Terraboard config file:

```yaml
web:
  port: 8080
  oidc:
    enabled: true
    issuer-url: "https://your-dex-server.example.com"
    client-id: "terraboard-ui"
    client-secret: "your-client-secret"
    redirect-url: "https://your-terraboard-url.example.com/callback"
```

### Environment Variables

Alternatively, you can use environment variables:

```bash
export TERRABOARD_OIDC_ENABLED=true
export TERRABOARD_OIDC_ISSUER_URL="https://your-dex-server.example.com"
export TERRABOARD_OIDC_CLIENT_ID="terraboard-ui"
export TERRABOARD_OIDC_CLIENT_SECRET="your-client-secret"
export TERRABOARD_OIDC_REDIRECT_URL="https://your-terraboard-url.example.com/callback"
```

### Dex Server Configuration

Configure your Dex server with a static client for Terraboard:

```yaml
issuer: https://your-dex-server.example.com

oauth2:
  skipApprovalScreen: true

staticClients:
  - id: terraboard-ui
    name: "Terraboard UI"
    redirectURIs:
      - "https://your-terraboard-url.example.com/callback"
    secret: "your-client-secret"

# Example connector configuration (adapt to your needs)
connectors:
  - type: oidc
    id: okta
    name: Okta
    config:
      issuer: "https://your-org.okta.com"
      clientID: "your-okta-client-id"
      clientSecret: "your-okta-client-secret"
      redirectURI: "https://your-dex-server.example.com/callback"
```

## Authentication Flow

1. **Unauthenticated Access**: When a user visits Terraboard without authentication, they see a login screen
2. **Login Initiation**: Clicking "Sign in with OIDC" redirects to `/auth/login`
3. **OIDC Redirect**: Terraboard redirects to the Dex authorization endpoint
4. **User Authentication**: User authenticates with Dex (via configured connectors)
5. **Authorization Code**: Dex redirects back to `/callback` with authorization code
6. **Token Exchange**: Terraboard exchanges the code for access and ID tokens
7. **Authentication Complete**: User is authenticated and can access Terraboard

## API Endpoints

The following authentication endpoints are available:

- `GET /auth/login` - Initiates OIDC login flow
- `GET /auth/callback` - Handles OIDC callback and token exchange
- `POST /auth/logout` - Logs out user and clears authentication cookies
- `GET /auth/status` - Returns current authentication status

## Security Considerations

- **HTTPS Required**: Use HTTPS in production for secure token transmission
- **State Validation**: CSRF protection via state parameter validation
- **Secure Cookies**: Authentication tokens stored in HttpOnly, Secure cookies
- **Token Validation**: ID tokens are verified using Dex's public keys

## Troubleshooting

### Common Issues

1. **"OIDC authentication is not enabled"**
   - Ensure `oidc.enabled: true` is set in configuration
   - Check environment variables are properly set

2. **"Failed to get OIDC provider"**
   - Verify the issuer URL is correct and accessible
   - Check network connectivity to Dex server
   - Ensure Dex is properly configured and running

3. **"Invalid state parameter"**
   - Usually indicates CSRF protection triggered
   - Clear browser cookies and try again
   - Check for clock synchronization issues

4. **"No ID token in response"**
   - Verify Dex client configuration includes `openid` scope
   - Check Dex logs for authentication errors

### Debug Mode

Enable debug logging to troubleshoot authentication issues:

```yaml
log:
  level: "debug"
```

Or via environment variable:
```bash
export TERRABOARD_LOG_LEVEL=debug
```

## Migration from Proxy Authentication

If you're currently using proxy-based authentication (oauth2_proxy, nginx auth, etc.), you can:

1. **Gradual Migration**: Keep proxy authentication running while testing OIDC
2. **Parallel Operation**: Both authentication methods can work simultaneously
3. **Full Migration**: Once OIDC is tested, disable proxy authentication

The `/api/user` endpoint automatically detects the authentication method and returns appropriate user information.

## Example Deployment

Here's a complete example using Docker Compose:

```yaml
version: '3.8'
services:
  terraboard:
    image: camptocamp/terraboard:latest
    ports:
      - "8080:8080"
    environment:
      - TERRABOARD_OIDC_ENABLED=true
      - TERRABOARD_OIDC_ISSUER_URL=https://dex.example.com
      - TERRABOARD_OIDC_CLIENT_ID=terraboard-ui
      - TERRABOARD_OIDC_CLIENT_SECRET=secret
      - TERRABOARD_OIDC_REDIRECT_URL=https://terraboard.example.com/callback
      - DB_HOST=postgres
      - DB_USER=terraboard
      - DB_PASSWORD=password
      - AWS_ACCESS_KEY_ID=your-key
      - AWS_SECRET_ACCESS_KEY=your-secret
      - AWS_BUCKET=your-bucket
    depends_on:
      - postgres

  postgres:
    image: postgres:13
    environment:
      - POSTGRES_USER=terraboard
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=terraboard
```

## Support

For issues with OIDC authentication:

1. Check Terraboard logs for authentication errors
2. Verify Dex server configuration and logs
3. Test OIDC flow manually using curl or similar tools
4. Open an issue with configuration details (redact secrets)