package auth

import (
	"reflect"
	"testing"

	"github.com/camptocamp/terraboard/config"
)

func TestSetup_simple(t *testing.T) {
	expected := "/log/me/out"

	c := config.Config{}
	c.Web.LogoutURL = expected

	Setup(&c)

	if logoutURL != expected {
		t.Fatalf("Expected %s, got %s", expected, logoutURL)
	}
}

func TestUserInfo(t *testing.T) {
	expected := User{
		Name:          "foo",
		LogoutURL:     "/log/me/out",
		AvatarURL:     "http://www.gravatar.com/avatar/b48def645758b95537d4424c84d1a9ff",
		IsOIDC:        false,
		Authenticated: true,
	}

	u := UserInfo("foo", "foo@example.com")

	if !reflect.DeepEqual(u, expected) {
		t.Fatalf("Expected %v, got %v", expected, u)
	}
}

func TestOIDCSetupDisabled(t *testing.T) {
	config := &config.OIDCConfig{
		Enabled: false,
	}

	err := SetupOIDC(config)
	if err != nil {
		t.Errorf("Expected no error when OIDC is disabled, got: %v", err)
	}

	if IsOIDCEnabled() {
		t.Error("Expected OIDC to be disabled")
	}
}

func TestOIDCSetupIncompleteConfig(t *testing.T) {
	config := &config.OIDCConfig{
		Enabled:   true,
		IssuerURL: "https://example.com",
		// Missing other required fields
	}

	err := SetupOIDC(config)
	if err == nil {
		t.Error("Expected error with incomplete OIDC config")
	}
}

func TestGenerateStateString(t *testing.T) {
	state1, err1 := GenerateStateString()
	if err1 != nil {
		t.Errorf("Error generating state string: %v", err1)
	}

	state2, err2 := GenerateStateString()
	if err2 != nil {
		t.Errorf("Error generating second state string: %v", err2)
	}

	if state1 == state2 {
		t.Error("Generated state strings should be unique")
	}

	if len(state1) == 0 {
		t.Error("State string should not be empty")
	}
}

func TestOIDCUserInfo(t *testing.T) {
	claims := map[string]interface{}{
		"name":  "OIDC User",
		"email": "oidc@example.com",
	}

	user := OIDCUserInfo(claims)
	if user.Name != "OIDC User" {
		t.Errorf("Expected name 'OIDC User', got '%s'", user.Name)
	}
	if !user.Authenticated {
		t.Error("Expected user to be authenticated")
	}
	if !user.IsOIDC {
		t.Error("Expected user to be OIDC authenticated")
	}
	if user.LogoutURL != "/auth/logout" {
		t.Errorf("Expected logout URL '/auth/logout', got '%s'", user.LogoutURL)
	}
}
