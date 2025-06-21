package auth

import (
	"crypto/md5"
	"fmt"

	"github.com/camptocamp/terraboard/config"
)

var logoutURL string

// User is an authenticated user
type User struct {
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	LogoutURL    string `json:"logout_url"`
	IsOIDC       bool   `json:"is_oidc"`
	Authenticated bool   `json:"authenticated"`
}

// Setup sets up authentication
func Setup(c *config.Config) {
	logoutURL = c.Web.LogoutURL
	
	// Setup OIDC if enabled
	if err := SetupOIDC(&c.Web.OIDC); err != nil {
		// Log error but don't fail startup - fall back to proxy auth
		fmt.Printf("Failed to setup OIDC: %v\n", err)
	}
}

// UserInfo returns a User given a name and email (for proxy-based auth)
func UserInfo(name, email string) (user User) {
	user = User{
		LogoutURL: logoutURL,
		IsOIDC:    false,
	}

	if email != "" {
		user.Name = name
		user.AvatarURL = fmt.Sprintf("http://www.gravatar.com/avatar/%x", md5.Sum([]byte(email)))
		user.Authenticated = true
	}

	return
}

// OIDCUserInfo returns a User from OIDC claims
func OIDCUserInfo(claims map[string]interface{}) User {
	user := User{
		LogoutURL:     "/auth/logout",
		IsOIDC:        true,
		Authenticated: true,
	}

	if name, ok := claims["name"].(string); ok {
		user.Name = name
	} else if preferredUsername, ok := claims["preferred_username"].(string); ok {
		user.Name = preferredUsername
	}

	if email, ok := claims["email"].(string); ok && email != "" {
		user.AvatarURL = fmt.Sprintf("http://www.gravatar.com/avatar/%x", md5.Sum([]byte(email)))
	}

	return user
}
