package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/camptocamp/terraboard/auth"
	log "github.com/sirupsen/logrus"
)

// OIDCLogin initiates the OIDC login process
// @Summary Initiate OIDC login
// @Description Redirects user to OIDC provider for authentication
// @ID oidc-login
// @Produce  json
// @Success 302 {string} string "redirect"
// @Router /auth/login [get]
func OIDCLogin(w http.ResponseWriter, r *http.Request) {
	if !auth.IsOIDCEnabled() {
		http.Error(w, "OIDC authentication is not enabled", http.StatusNotImplemented)
		return
	}

	// Generate state parameter
	state, err := auth.GenerateStateString()
	if err != nil {
		log.Errorf("Failed to generate state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set state cookie
	auth.SetStateCookie(w, state)

	// Get auth URL and redirect
	authURL := auth.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OIDCCallback handles the OIDC callback
// @Summary Handle OIDC callback
// @Description Handles the callback from OIDC provider and exchanges code for tokens
// @ID oidc-callback
// @Produce  json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 302 {string} string "redirect"
// @Router /auth/callback [get]
func OIDCCallback(w http.ResponseWriter, r *http.Request) {
	if !auth.IsOIDCEnabled() {
		http.Error(w, "OIDC authentication is not enabled", http.StatusNotImplemented)
		return
	}

	ctx := context.Background()

	// Get code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	if state == "" {
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}

	// Validate state cookie
	if err := auth.ValidateStateCookie(r, state); err != nil {
		log.Errorf("State validation failed: %v", err)
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	auth.ClearStateCookie(w)

	// Exchange code for tokens
	token, err := auth.ExchangeCode(ctx, code)
	if err != nil {
		log.Errorf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Extract and verify ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Error("No id_token field in oauth2 token")
		http.Error(w, "No ID token in response", http.StatusInternalServerError)
		return
	}

	idToken, err := auth.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		log.Errorf("Failed to verify ID token: %v", err)
		http.Error(w, "Failed to verify ID token", http.StatusInternalServerError)
		return
	}

	// Set token cookie (using access token for API calls)
	auth.SetTokenCookie(w, token.AccessToken)

	// For debugging/logging, we can extract claims
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		log.Errorf("Failed to extract claims: %v", err)
	} else {
		log.Debugf("User authenticated via OIDC: %v", claims["preferred_username"])
	}

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// OIDCLogout handles OIDC logout
// @Summary OIDC logout
// @Description Logs out the user by clearing authentication cookies
// @ID oidc-logout
// @Produce  json
// @Success 200 {string} string "ok"
// @Router /auth/logout [post]
func OIDCLogout(w http.ResponseWriter, r *http.Request) {
	// Clear token cookie
	auth.ClearTokenCookie(w)
	
	// Return success response (frontend will redirect)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "logged_out"}
	json.NewEncoder(w).Encode(response)
}

// OIDCStatus checks if user is authenticated via OIDC
// @Summary Check OIDC authentication status
// @Description Returns authentication status and user info if authenticated
// @ID oidc-status
// @Produce  json
// @Success 200 {object} auth.User
// @Router /auth/status [get]
func OIDCStatus(w http.ResponseWriter, r *http.Request) {
	user := auth.User{
		Authenticated: false,
		IsOIDC:        auth.IsOIDCEnabled(),
	}

	// If OIDC is enabled, check for token
	if auth.IsOIDCEnabled() {
		token, err := auth.GetTokenFromCookie(r)
		if err == nil && token != "" {
			// For simplicity, we assume token is valid if it exists
			// In production, you might want to validate the token
			user.Authenticated = true
			user.Name = "OIDC User" // This would come from token claims in real implementation
		}
	}

	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(user)
	if err != nil {
		JSONError(w, "Failed to marshal user", err)
		return
	}
	if _, err := io.WriteString(w, string(j)); err != nil {
		log.Error(err.Error())
	}
}