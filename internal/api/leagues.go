package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// generateInviteCode generates a random 8-character invite code
func generateInviteCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// handleGetLeagues returns all leagues for the current user
func (s *Server) handleGetLeagues(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT l.id, l.name, l.invite_code, l.commissioner_id, l.max_members, l.scoring_type, l.created_at,
		       (SELECT COUNT(*) FROM league_members WHERE league_id = l.id) as member_count
		FROM leagues l
		INNER JOIN league_members lm ON l.id = lm.league_id
		WHERE lm.user_id = $1
		ORDER BY l.created_at DESC
	`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch leagues")
		return
	}
	defer rows.Close()

	var leagues []map[string]interface{}
	for rows.Next() {
		var id, commissionerID uuid.UUID
		var name, inviteCode, scoringType string
		var maxMembers int
		var memberCount int64
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &inviteCode, &commissionerID, &maxMembers, &scoringType, &createdAt, &memberCount); err != nil {
			continue
		}

		leagues = append(leagues, map[string]interface{}{
			"id":              id,
			"name":            name,
			"invite_code":     inviteCode,
			"commissioner_id": commissionerID,
			"max_members":     maxMembers,
			"scoring_type":    scoringType,
			"member_count":    memberCount,
			"created_at":      createdAt,
		})
	}

	if leagues == nil {
		leagues = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, leagues)
}

// handleCreateLeague creates a new league
func (s *Server) handleCreateLeague(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req struct {
		Name        string `json:"name"`
		TeamName    string `json:"team_name"`
		MaxMembers  int    `json:"max_members"`
		ScoringType string `json:"scoring_type"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.TeamName == "" {
		writeError(w, http.StatusBadRequest, "name and team_name are required")
		return
	}

	// Set defaults
	if req.MaxMembers == 0 {
		req.MaxMembers = 10
	}
	if req.ScoringType == "" {
		req.ScoringType = "ppr"
	}

	ctx := r.Context()
	inviteCode := generateInviteCode()

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback(ctx)

	// Create league
	var leagueID uuid.UUID
	var createdAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO leagues (name, invite_code, commissioner_id, max_members, scoring_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, req.Name, inviteCode, userID, req.MaxMembers, req.ScoringType).Scan(&leagueID, &createdAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create league")
		return
	}

	// Add commissioner as first member
	_, err = tx.Exec(ctx, `
		INSERT INTO league_members (league_id, user_id, team_name)
		VALUES ($1, $2, $3)
	`, leagueID, userID, req.TeamName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add member")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":              leagueID,
		"name":            req.Name,
		"invite_code":     inviteCode,
		"commissioner_id": userID,
		"max_members":     req.MaxMembers,
		"scoring_type":    req.ScoringType,
		"member_count":    1,
		"created_at":      createdAt,
	})
}

// handleJoinLeague joins a league via invite code
func (s *Server) handleJoinLeague(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req struct {
		InviteCode string `json:"invite_code"`
		TeamName   string `json:"team_name"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.InviteCode == "" || req.TeamName == "" {
		writeError(w, http.StatusBadRequest, "invite_code and team_name are required")
		return
	}

	ctx := r.Context()

	// Find league by invite code
	var leagueID uuid.UUID
	var maxMembers int
	err := s.pool.QueryRow(ctx, `
		SELECT id, max_members FROM leagues WHERE invite_code = $1
	`, req.InviteCode).Scan(&leagueID, &maxMembers)
	if err != nil {
		writeError(w, http.StatusNotFound, "invalid invite code")
		return
	}

	// Check if already a member
	var existingCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM league_members WHERE league_id = $1 AND user_id = $2
	`, leagueID, userID).Scan(&existingCount)
	if existingCount > 0 {
		writeError(w, http.StatusConflict, "already a member of this league")
		return
	}

	// Check member count
	var currentCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM league_members WHERE league_id = $1
	`, leagueID).Scan(&currentCount)
	if currentCount >= maxMembers {
		writeError(w, http.StatusConflict, "league is full")
		return
	}

	// Add member
	var memberID uuid.UUID
	var joinedAt time.Time
	err = s.pool.QueryRow(ctx, `
		INSERT INTO league_members (league_id, user_id, team_name)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at
	`, leagueID, userID, req.TeamName).Scan(&memberID, &joinedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to join league")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        memberID,
		"league_id": leagueID,
		"user_id":   userID,
		"team_name": req.TeamName,
		"joined_at": joinedAt,
	})
}

// handleGetLeague returns a specific league
func (s *Server) handleGetLeague(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	ctx := r.Context()

	// Verify membership
	var memberCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM league_members WHERE league_id = $1 AND user_id = $2
	`, leagueID, userID).Scan(&memberCount)
	if memberCount == 0 {
		writeError(w, http.StatusForbidden, "not a member of this league")
		return
	}

	// Get league details
	var id, commissionerID uuid.UUID
	var name, inviteCode, scoringType string
	var maxMembers int
	var totalMembers int64
	var createdAt time.Time

	err = s.pool.QueryRow(ctx, `
		SELECT l.id, l.name, l.invite_code, l.commissioner_id, l.max_members, l.scoring_type, l.created_at,
		       (SELECT COUNT(*) FROM league_members WHERE league_id = l.id) as member_count
		FROM leagues l
		WHERE l.id = $1
	`, leagueID).Scan(&id, &name, &inviteCode, &commissionerID, &maxMembers, &scoringType, &createdAt, &totalMembers)
	if err != nil {
		writeError(w, http.StatusNotFound, "league not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":              id,
		"name":            name,
		"invite_code":     inviteCode,
		"commissioner_id": commissionerID,
		"max_members":     maxMembers,
		"scoring_type":    scoringType,
		"member_count":    totalMembers,
		"created_at":      createdAt,
	})
}

// handleUpdateLeague updates a league (commissioner only)
func (s *Server) handleUpdateLeague(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		MaxMembers  *int    `json:"max_members"`
		ScoringType *string `json:"scoring_type"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()

	// Update league (only if commissioner)
	result, err := s.pool.Exec(ctx, `
		UPDATE leagues
		SET name = COALESCE($2, name),
		    max_members = COALESCE($3, max_members),
		    scoring_type = COALESCE($4, scoring_type)
		WHERE id = $1 AND commissioner_id = $5
	`, leagueID, req.Name, req.MaxMembers, req.ScoringType, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update league")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusForbidden, "not authorized to update this league")
		return
	}

	// Return updated league
	s.handleGetLeague(w, r)
}

// handleDeleteLeague deletes a league (commissioner only)
func (s *Server) handleDeleteLeague(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	result, err := s.pool.Exec(r.Context(), `
		DELETE FROM leagues WHERE id = $1 AND commissioner_id = $2
	`, leagueID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete league")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusForbidden, "not authorized to delete this league")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetLeagueMembers returns all members of a league
func (s *Server) handleGetLeagueMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	ctx := r.Context()

	// Verify membership
	var memberCheck int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM league_members WHERE league_id = $1 AND user_id = $2
	`, leagueID, userID).Scan(&memberCheck)
	if memberCheck == 0 {
		writeError(w, http.StatusForbidden, "not a member of this league")
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT lm.id, lm.league_id, lm.user_id, lm.team_name, lm.joined_at,
		       u.display_name, u.avatar_url
		FROM league_members lm
		INNER JOIN users u ON lm.user_id = u.id
		WHERE lm.league_id = $1
		ORDER BY lm.joined_at
	`, leagueID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch members")
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var id, leagueIDResult, memberUserID uuid.UUID
		var teamName, displayName string
		var avatarURL *string
		var joinedAt time.Time

		if err := rows.Scan(&id, &leagueIDResult, &memberUserID, &teamName, &joinedAt, &displayName, &avatarURL); err != nil {
			continue
		}

		members = append(members, map[string]interface{}{
			"id":           id,
			"league_id":    leagueIDResult,
			"user_id":      memberUserID,
			"team_name":    teamName,
			"joined_at":    joinedAt,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
	}

	if members == nil {
		members = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, members)
}

// handleGetLeagueStandings returns current standings for the active season
func (s *Server) handleGetLeagueStandings(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	ctx := r.Context()

	// Verify membership
	var memberCheck int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM league_members WHERE league_id = $1 AND user_id = $2
	`, leagueID, userID).Scan(&memberCheck)
	if memberCheck == 0 {
		writeError(w, http.StatusForbidden, "not a member of this league")
		return
	}

	// Get active season
	var seasonID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id FROM seasons WHERE league_id = $1 AND status = 'active'
	`, leagueID).Scan(&seasonID)
	if err != nil {
		writeJSON(w, http.StatusOK, []map[string]interface{}{})
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT s.id, s.season_id, s.member_id, s.weekly_wins, s.total_points, s.best_week, s.weeks_played, s.current_rank,
		       lm.team_name, u.display_name, u.avatar_url
		FROM standings s
		INNER JOIN league_members lm ON s.member_id = lm.id
		INNER JOIN users u ON lm.user_id = u.id
		WHERE s.season_id = $1
		ORDER BY s.total_points DESC
	`, seasonID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch standings")
		return
	}
	defer rows.Close()

	var standings []map[string]interface{}
	for rows.Next() {
		var id, sID, memberID uuid.UUID
		var weeklyWins, weeksPlayed, currentRank int
		var totalPoints, bestWeek float64
		var teamName, displayName string
		var avatarURL *string

		if err := rows.Scan(&id, &sID, &memberID, &weeklyWins, &totalPoints, &bestWeek, &weeksPlayed, &currentRank, &teamName, &displayName, &avatarURL); err != nil {
			continue
		}

		standings = append(standings, map[string]interface{}{
			"id":           id,
			"season_id":    sID,
			"member_id":    memberID,
			"weekly_wins":  weeklyWins,
			"total_points": totalPoints,
			"best_week":    bestWeek,
			"weeks_played": weeksPlayed,
			"current_rank": currentRank,
			"team_name":    teamName,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
	}

	if standings == nil {
		standings = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, standings)
}

// Season handlers

// handleGetSeasons returns all seasons for a league
func (s *Server) handleGetSeasons(w http.ResponseWriter, r *http.Request) {
	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT id, league_id, year, current_week, status, champion_member_id, created_at
		FROM seasons WHERE league_id = $1
		ORDER BY year DESC
	`, leagueID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch seasons")
		return
	}
	defer rows.Close()

	var seasons []map[string]interface{}
	for rows.Next() {
		var id, leagueIDResult uuid.UUID
		var championMemberID *uuid.UUID
		var year, currentWeek int
		var status string
		var createdAt time.Time

		if err := rows.Scan(&id, &leagueIDResult, &year, &currentWeek, &status, &championMemberID, &createdAt); err != nil {
			continue
		}

		seasons = append(seasons, map[string]interface{}{
			"id":                 id,
			"league_id":          leagueIDResult,
			"year":               year,
			"current_week":       currentWeek,
			"status":             status,
			"champion_member_id": championMemberID,
			"created_at":         createdAt,
		})
	}

	if seasons == nil {
		seasons = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, seasons)
}

// handleCreateSeason creates a new season for a league
func (s *Server) handleCreateSeason(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	var req struct {
		Year int `json:"year"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Year == 0 {
		req.Year = time.Now().Year()
	}

	ctx := r.Context()

	// Verify commissioner
	var commissionerID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT commissioner_id FROM leagues WHERE id = $1
	`, leagueID).Scan(&commissionerID)
	if err != nil || commissionerID != userID {
		writeError(w, http.StatusForbidden, "only commissioner can create seasons")
		return
	}

	// Create season
	var seasonID uuid.UUID
	var createdAt time.Time
	err = s.pool.QueryRow(ctx, `
		INSERT INTO seasons (league_id, year, current_week, status)
		VALUES ($1, $2, 1, 'active')
		RETURNING id, created_at
	`, leagueID, req.Year).Scan(&seasonID, &createdAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create season")
		return
	}

	// Initialize standings for all members
	_, err = s.pool.Exec(ctx, `
		INSERT INTO standings (season_id, member_id, weekly_wins, total_points, best_week, weeks_played, current_rank)
		SELECT $1, id, 0, 0, 0, 0, 0
		FROM league_members WHERE league_id = $2
	`, seasonID, leagueID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initialize standings")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           seasonID,
		"league_id":    leagueID,
		"year":         req.Year,
		"current_week": 1,
		"status":       "active",
		"created_at":   createdAt,
	})
}

// handleGetActiveSeason returns the active season for a league
func (s *Server) handleGetActiveSeason(w http.ResponseWriter, r *http.Request) {
	leagueID, err := uuid.Parse(chi.URLParam(r, "leagueID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid league ID")
		return
	}

	var id, leagueIDResult uuid.UUID
	var championMemberID *uuid.UUID
	var year, currentWeek int
	var status string
	var createdAt time.Time

	err = s.pool.QueryRow(r.Context(), `
		SELECT id, league_id, year, current_week, status, champion_member_id, created_at
		FROM seasons WHERE league_id = $1 AND status = 'active'
	`, leagueID).Scan(&id, &leagueIDResult, &year, &currentWeek, &status, &championMemberID, &createdAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "no active season found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":                 id,
		"league_id":          leagueIDResult,
		"year":               year,
		"current_week":       currentWeek,
		"status":             status,
		"champion_member_id": championMemberID,
		"created_at":         createdAt,
	})
}

// handleGetSeason returns a specific season
func (s *Server) handleGetSeason(w http.ResponseWriter, r *http.Request) {
	seasonID, err := uuid.Parse(chi.URLParam(r, "seasonID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid season ID")
		return
	}

	var id, leagueID uuid.UUID
	var championMemberID *uuid.UUID
	var year, currentWeek int
	var status string
	var createdAt time.Time

	err = s.pool.QueryRow(r.Context(), `
		SELECT id, league_id, year, current_week, status, champion_member_id, created_at
		FROM seasons WHERE id = $1
	`, seasonID).Scan(&id, &leagueID, &year, &currentWeek, &status, &championMemberID, &createdAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "season not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":                 id,
		"league_id":          leagueID,
		"year":               year,
		"current_week":       currentWeek,
		"status":             status,
		"champion_member_id": championMemberID,
		"created_at":         createdAt,
	})
}

// handleGetSeasonStandings returns standings for a specific season
func (s *Server) handleGetSeasonStandings(w http.ResponseWriter, r *http.Request) {
	seasonID, err := uuid.Parse(chi.URLParam(r, "seasonID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid season ID")
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT s.id, s.season_id, s.member_id, s.weekly_wins, s.total_points, s.best_week, s.weeks_played, s.current_rank,
		       lm.team_name, u.display_name, u.avatar_url
		FROM standings s
		INNER JOIN league_members lm ON s.member_id = lm.id
		INNER JOIN users u ON lm.user_id = u.id
		WHERE s.season_id = $1
		ORDER BY s.total_points DESC
	`, seasonID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch standings")
		return
	}
	defer rows.Close()

	var standings []map[string]interface{}
	for rows.Next() {
		var id, sID, memberID uuid.UUID
		var weeklyWins, weeksPlayed, currentRank int
		var totalPoints, bestWeek float64
		var teamName, displayName string
		var avatarURL *string

		if err := rows.Scan(&id, &sID, &memberID, &weeklyWins, &totalPoints, &bestWeek, &weeksPlayed, &currentRank, &teamName, &displayName, &avatarURL); err != nil {
			continue
		}

		standings = append(standings, map[string]interface{}{
			"id":           id,
			"season_id":    sID,
			"member_id":    memberID,
			"weekly_wins":  weeklyWins,
			"total_points": totalPoints,
			"best_week":    bestWeek,
			"weeks_played": weeksPlayed,
			"current_rank": currentRank,
			"team_name":    teamName,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
	}

	if standings == nil {
		standings = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, standings)
}

// Draft handlers (stubs for Phase 2)

// handleGetWeeklyDraft returns draft status for a week
func (s *Server) handleGetWeeklyDraft(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	seasonID, err := uuid.Parse(chi.URLParam(r, "seasonID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid season ID")
		return
	}

	week, err := strconv.Atoi(chi.URLParam(r, "week"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid week")
		return
	}

	ctx := r.Context()

	// Get the user's member ID for this league
	var memberID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT lm.id FROM league_members lm
		INNER JOIN seasons s ON s.league_id = lm.league_id
		WHERE s.id = $1 AND lm.user_id = $2
	`, seasonID, userID).Scan(&memberID)
	if err != nil {
		writeError(w, http.StatusForbidden, "not a member of this league")
		return
	}

	// Get draft status
	var draftID uuid.UUID
	var status string
	var startedAt, completedAt *time.Time

	err = s.pool.QueryRow(ctx, `
		SELECT id, status, started_at, completed_at
		FROM weekly_drafts
		WHERE season_id = $1 AND member_id = $2 AND week = $3
	`, seasonID, memberID, week).Scan(&draftID, &status, &startedAt, &completedAt)

	if err != nil {
		// No draft exists yet
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"season_id": seasonID,
			"member_id": memberID,
			"week":      week,
			"status":    "not_started",
		})
		return
	}

	// Get picks if any
	rows, err := s.pool.Query(ctx, `
		SELECT id, pick_number, nfl_player_id, player_name, position, team_drawn, picked_at
		FROM draft_picks WHERE weekly_draft_id = $1 ORDER BY pick_number
	`, draftID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch picks")
		return
	}
	defer rows.Close()

	var picks []map[string]interface{}
	for rows.Next() {
		var pickID uuid.UUID
		var pickNumber int
		var nflPlayerID, playerName, position, teamDrawn string
		var pickedAt time.Time

		if err := rows.Scan(&pickID, &pickNumber, &nflPlayerID, &playerName, &position, &teamDrawn, &pickedAt); err != nil {
			continue
		}

		picks = append(picks, map[string]interface{}{
			"id":            pickID,
			"pick_number":   pickNumber,
			"nfl_player_id": nflPlayerID,
			"player_name":   playerName,
			"position":      position,
			"team_drawn":    teamDrawn,
			"picked_at":     pickedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           draftID,
		"season_id":    seasonID,
		"member_id":    memberID,
		"week":         week,
		"status":       status,
		"started_at":   startedAt,
		"completed_at": completedAt,
		"picks":        picks,
	})
}

// handleStartWeeklyDraft starts or resumes a weekly draft
func (s *Server) handleStartWeeklyDraft(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	seasonID, err := uuid.Parse(chi.URLParam(r, "seasonID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid season ID")
		return
	}

	week, err := strconv.Atoi(chi.URLParam(r, "week"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid week")
		return
	}

	ctx := r.Context()

	// Get the user's member ID for this league
	var memberID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT lm.id FROM league_members lm
		INNER JOIN seasons s ON s.league_id = lm.league_id
		WHERE s.id = $1 AND lm.user_id = $2
	`, seasonID, userID).Scan(&memberID)
	if err != nil {
		writeError(w, http.StatusForbidden, "not a member of this league")
		return
	}

	// Check if draft already exists
	var existingID uuid.UUID
	var existingStatus string
	err = s.pool.QueryRow(ctx, `
		SELECT id, status FROM weekly_drafts
		WHERE season_id = $1 AND member_id = $2 AND week = $3
	`, seasonID, memberID, week).Scan(&existingID, &existingStatus)

	if err == nil {
		// Draft exists
		if existingStatus == "complete" {
			writeError(w, http.StatusConflict, "draft already completed")
			return
		}
		// Return existing draft
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":        existingID,
			"season_id": seasonID,
			"member_id": memberID,
			"week":      week,
			"status":    existingStatus,
			"message":   "draft resumed",
		})
		return
	}

	// Create new draft
	var draftID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		INSERT INTO weekly_drafts (season_id, member_id, week, status, started_at)
		VALUES ($1, $2, $3, 'in_progress', NOW())
		RETURNING id
	`, seasonID, memberID, week).Scan(&draftID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create draft")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        draftID,
		"season_id": seasonID,
		"member_id": memberID,
		"week":      week,
		"status":    "in_progress",
		"message":   "draft started",
	})
}
