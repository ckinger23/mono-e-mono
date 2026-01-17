package auth

import (
	"fmt"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	CallbackBaseURL    string
}

// OAuthManager handles OAuth authentication
type OAuthManager struct {
	config *OAuthConfig
}

// NewOAuthManager creates a new OAuth manager and initializes providers
func NewOAuthManager(config *OAuthConfig) *OAuthManager {
	providers := []goth.Provider{}

	if config.GoogleClientID != "" && config.GoogleClientSecret != "" {
		providers = append(providers, google.New(
			config.GoogleClientID,
			config.GoogleClientSecret,
			fmt.Sprintf("%s/api/auth/google/callback", config.CallbackBaseURL),
			"email", "profile",
		))
	}

	if config.GitHubClientID != "" && config.GitHubClientSecret != "" {
		providers = append(providers, github.New(
			config.GitHubClientID,
			config.GitHubClientSecret,
			fmt.Sprintf("%s/api/auth/github/callback", config.CallbackBaseURL),
			"user:email",
		))
	}

	if len(providers) > 0 {
		goth.UseProviders(providers...)
	}

	return &OAuthManager{config: config}
}

// BeginAuth starts the OAuth flow for a provider
func (m *OAuthManager) BeginAuth(w http.ResponseWriter, r *http.Request, provider string) error {
	// Set the provider in the request context for gothic
	q := r.URL.Query()
	q.Set("provider", provider)
	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
	return nil
}

// CompleteAuth completes the OAuth flow and returns the user info
func (m *OAuthManager) CompleteAuth(w http.ResponseWriter, r *http.Request) (*OAuthUser, error) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return nil, fmt.Errorf("failed to complete OAuth: %w", err)
	}

	return &OAuthUser{
		Provider:    user.Provider,
		ProviderID:  user.UserID,
		Email:       user.Email,
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		AccessToken: user.AccessToken,
	}, nil
}

// OAuthUser represents a user authenticated via OAuth
type OAuthUser struct {
	Provider    string
	ProviderID  string
	Email       string
	Name        string
	AvatarURL   string
	AccessToken string
}

// GetProviderFromRequest extracts the provider name from the request
func GetProviderFromRequest(r *http.Request) string {
	return r.URL.Query().Get("provider")
}
