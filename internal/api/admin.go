package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ckinger23/mono-e-mono/internal/nfl"
	"github.com/go-chi/chi/v5"
)

// handleSyncPlayers syncs all NFL players from Sleeper API
func (s *Server) handleSyncPlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := slog.Default()
	syncer := nfl.NewPlayerSyncer(s.pool, logger)

	result, err := syncer.SyncAllPlayers(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to sync players: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handleSyncWeeklyStats syncs stats for a specific week from Sleeper API
func (s *Server) handleSyncWeeklyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	yearStr := r.URL.Query().Get("year")
	weekStr := r.URL.Query().Get("week")

	if yearStr == "" || weekStr == "" {
		writeError(w, http.StatusBadRequest, "year and week query parameters are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid year")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid week")
		return
	}

	logger := slog.Default()
	syncer := nfl.NewStatsSyncer(s.pool, logger)

	result, err := syncer.SyncWeeklyStats(ctx, year, week)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to sync stats: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handleCalculateScores calculates and stores scores for a week in a season
func (s *Server) handleCalculateScores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seasonID := chi.URLParam(r, "seasonID")
	weekStr := r.URL.Query().Get("week")

	if weekStr == "" {
		writeError(w, http.StatusBadRequest, "week query parameter is required")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid week")
		return
	}

	logger := slog.Default()
	scoringService := nfl.NewScoringService(s.pool, logger)

	result, err := scoringService.CalculateWeeklyScores(ctx, seasonID, week)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to calculate scores: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handleRecalculateAllWeeks recalculates scores for all weeks in a season
func (s *Server) handleRecalculateAllWeeks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seasonID := chi.URLParam(r, "seasonID")

	logger := slog.Default()
	scoringService := nfl.NewScoringService(s.pool, logger)

	if err := scoringService.RecalculateAllWeeks(ctx, seasonID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to recalculate scores: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// handleGetNFLState returns the current NFL state (week, season)
func (s *Server) handleGetNFLState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client := nfl.NewClient()
	state, err := client.GetNFLState(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get NFL state: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, state)
}

// handleGetWeeklyLeaders returns top scoring players for a week
func (s *Server) handleGetWeeklyLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	yearStr := r.URL.Query().Get("year")
	weekStr := r.URL.Query().Get("week")
	limitStr := r.URL.Query().Get("limit")

	if yearStr == "" || weekStr == "" {
		writeError(w, http.StatusBadRequest, "year and week query parameters are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid year")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid week")
		return
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logger := slog.Default()
	syncer := nfl.NewStatsSyncer(s.pool, logger)

	leaders, err := syncer.GetWeeklyLeaders(ctx, year, week, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get leaders: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, leaders)
}

// handleGetNFLPlayers returns NFL players, optionally filtered by team or position
func (s *Server) handleGetNFLPlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	team := r.URL.Query().Get("team")
	position := r.URL.Query().Get("position")

	logger := slog.Default()
	syncer := nfl.NewPlayerSyncer(s.pool, logger)

	var players []nfl.DBPlayer
	var err error

	if team != "" {
		players, err = syncer.GetPlayersByTeam(ctx, team)
	} else if position != "" {
		players, err = syncer.GetPlayersByPosition(ctx, position)
	} else {
		// Return empty list if no filter specified (too many to return all)
		writeJSON(w, http.StatusOK, []nfl.DBPlayer{})
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get players: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, players)
}

// handleGetNFLTeams returns all NFL teams that have players
func (s *Server) handleGetNFLTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := slog.Default()
	syncer := nfl.NewPlayerSyncer(s.pool, logger)

	teams, err := syncer.GetAllTeams(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get teams: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, teams)
}
