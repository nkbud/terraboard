package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/camptocamp/terraboard/config"
	"github.com/coreos/go-oidc/v3/oidc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// OIDCManager handles OIDC authentication
type OIDCManager struct {
	provider     *oidc.Provider
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	enabled      bool
}

var oidcManager *OIDCManager

// SetupOIDC initializes the OIDC provider and configuration
func SetupOIDC(config *config.OIDCConfig) error {
	if !config.Enabled {
		log.Info("OIDC authentication is disabled")
		oidcManager = &OIDCManager{enabled: false}
		return nil
	}

	if config.IssuerURL == "" || config.ClientID == "" || config.ClientSecret == "" || config.RedirectURL == "" {
		return fmt.Errorf("OIDC configuration incomplete: issuer-url, client-id, client-secret, and redirect-url are required")
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return fmt.Errorf("failed to get OIDC provider: %v", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: config.ClientID,
	}
	verifier := provider.Verifier(oidcConfig)

	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	oidcManager = &OIDCManager{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		enabled:      true,
	}

	log.Info("OIDC authentication is enabled")
	return nil
}

// IsOIDCEnabled returns true if OIDC is enabled
func IsOIDCEnabled() bool {
	return oidcManager != nil && oidcManager.enabled
}

// GetAuthURL generates the authentication URL for OIDC login
func GetAuthURL(state string) string {
	if !IsOIDCEnabled() {
		return ""
	}
	return oidcManager.oauth2Config.AuthCodeURL(state)
}

// ExchangeCode exchanges authorization code for tokens
func ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	if !IsOIDCEnabled() {
		return nil, fmt.Errorf("OIDC is not enabled")
	}
	return oidcManager.oauth2Config.Exchange(ctx, code)
}

// VerifyIDToken verifies and parses the ID token
func VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	if !IsOIDCEnabled() {
		return nil, fmt.Errorf("OIDC is not enabled")
	}
	return oidcManager.verifier.Verify(ctx, rawIDToken)
}

// GenerateStateString generates a random state string for OIDC
func GenerateStateString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// SetStateCookie sets a secure state cookie
func SetStateCookie(w http.ResponseWriter, state string) {
	cookie := &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		MaxAge:   int(10 * time.Minute.Seconds()), // 10 minutes
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

// ValidateStateCookie validates the state cookie and clears it
func ValidateStateCookie(r *http.Request, expectedState string) error {
	cookie, err := r.Cookie("oidc_state")
	if err != nil {
		return fmt.Errorf("missing state cookie: %v", err)
	}

	if cookie.Value != expectedState {
		return fmt.Errorf("state mismatch")
	}

	return nil
}

// ClearStateCookie clears the state cookie
func ClearStateCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

// SetTokenCookie sets the access token as a secure cookie
func SetTokenCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		MaxAge:   int(24 * time.Hour.Seconds()), // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

// ClearTokenCookie clears the access token cookie
func ClearTokenCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}

// GetTokenFromCookie retrieves the access token from cookie
func GetTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}