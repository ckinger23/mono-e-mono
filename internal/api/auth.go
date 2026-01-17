package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/ckinger23/mono-e-mono/internal/auth"
	"github.com/google/uuid"
)

// handleRegister handles user registration
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "email, password, and display_name are required")
		return
	}

	if err := auth.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	// Create user in database
	ctx := r.Context()
	user, err := s.createUser(ctx, req.Email, &passwordHash, req.DisplayName, nil, nil, nil)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

// handleLogin handles user login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Find user by email
	ctx := r.Context()
	user, err := s.getUserByEmail(ctx, req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check password
	if user.PasswordHash == nil || !auth.CheckPassword(req.Password, *user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":   user.ToResponse(),
		"tokens": tokens,
	})
}

// handleRefresh handles token refresh
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	// Generate new token pair
	tokens, err := s.jwtManager.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// handleOAuthGoogle initiates Google OAuth flow
func (s *Server) handleOAuthGoogle(w http.ResponseWriter, r *http.Request) {
	s.oauthManager.BeginAuth(w, r, "google")
}

// handleOAuthGoogleCallback handles Google OAuth callback
func (s *Server) handleOAuthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	s.handleOAuthCallback(w, r, "google")
}

// handleOAuthGitHub initiates GitHub OAuth flow
func (s *Server) handleOAuthGitHub(w http.ResponseWriter, r *http.Request) {
	s.oauthManager.BeginAuth(w, r, "github")
}

// handleOAuthGitHubCallback handles GitHub OAuth callback
func (s *Server) handleOAuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	s.handleOAuthCallback(w, r, "github")
}

func (s *Server) handleOAuthCallback(w http.ResponseWriter, r *http.Request, provider string) {
	oauthUser, err := s.oauthManager.CompleteAuth(w, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "OAuth authentication failed")
		return
	}

	ctx := r.Context()

	// Check if user exists with this OAuth provider
	user, err := s.getUserByOAuth(ctx, provider, oauthUser.ProviderID)
	if err != nil {
		// User doesn't exist, create new one
		user, err = s.createUser(ctx, oauthUser.Email, nil, oauthUser.Name, &oauthUser.AvatarURL, &provider, &oauthUser.ProviderID)
		if err != nil {
			// Check if email already exists
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
				writeError(w, http.StatusConflict, "email already registered with different method")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	// Redirect to frontend with tokens
	redirectURL := s.config.FrontendURL + "/auth/callback?access_token=" + tokens.AccessToken + "&refresh_token=" + tokens.RefreshToken
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// Database helper methods

func (s *Server) createUser(ctx context.Context, email string, passwordHash *string, displayName string, avatarURL, oauthProvider, oauthID *string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, avatar_url, oauth_provider, oauth_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, password_hash, display_name, avatar_url, oauth_provider, oauth_id, created_at, updated_at
	`, email, passwordHash, displayName, avatarURL, oauthProvider, oauthID)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.OAuthProvider, &user.OAuthID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Server) getUserByEmail(ctx context.Context, email string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, oauth_provider, oauth_id, created_at, updated_at
		FROM users WHERE email = $1
	`, email)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.OAuthProvider, &user.OAuthID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Server) getUserByOAuth(ctx context.Context, provider, providerID string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, oauth_provider, oauth_id, created_at, updated_at
		FROM users WHERE oauth_provider = $1 AND oauth_id = $2
	`, provider, providerID)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.OAuthProvider, &user.OAuthID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Server) getUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, avatar_url, oauth_provider, oauth_id, created_at, updated_at
		FROM users WHERE id = $1
	`, id)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.AvatarURL, &user.OAuthProvider, &user.OAuthID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// User type for internal use
type User struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  *string
	DisplayName   string
	AvatarURL     *string
	OAuthProvider *string
	OAuthID       *string
	CreatedAt     interface{}
	UpdatedAt     interface{}
}

func (u *User) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           u.ID,
		"email":        u.Email,
		"display_name": u.DisplayName,
		"avatar_url":   u.AvatarURL,
	}
}
