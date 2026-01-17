package api

import (
	"net/http"
)

// handleGetCurrentUser returns the current authenticated user
func (s *Server) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	user, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}

// handleUpdateCurrentUser updates the current authenticated user
func (s *Server) handleUpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req struct {
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()

	// Update user
	_, err := s.pool.Exec(ctx, `
		UPDATE users
		SET display_name = COALESCE($2, display_name),
		    avatar_url = COALESCE($3, avatar_url),
		    updated_at = NOW()
		WHERE id = $1
	`, userID, req.DisplayName, req.AvatarURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	// Fetch updated user
	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch updated user")
		return
	}

	writeJSON(w, http.StatusOK, user.ToResponse())
}
