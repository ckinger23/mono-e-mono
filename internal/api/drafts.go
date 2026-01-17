package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ckinger23/mono-e-mono/internal/game"
	"github.com/ckinger23/mono-e-mono/internal/models"
	"github.com/ckinger23/mono-e-mono/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, check against allowed origins
		return true
	},
}

// DraftHandler handles draft-related requests
type DraftHandler struct {
	pool         interface{ Query(context.Context, string, ...interface{}) (interface{ Close(); Next() bool; Scan(...interface{}) error }, error) }
	hub          *ws.Hub
	draftManager *game.WeeklyDraftManager
}

// handleDraftWebSocket handles WebSocket connections for drafts
func (s *Server) handleDraftWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get draft ID from URL
	draftIDStr := chi.URLParam(r, "draftID")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		http.Error(w, "invalid draft ID", http.StatusBadRequest)
		return
	}

	// Get token from query params (WebSocket can't use headers easily)
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Verify user owns this draft
	var memberID uuid.UUID
	var seasonID uuid.UUID
	var week int
	var status string
	err = s.pool.QueryRow(ctx, `
		SELECT wd.member_id, wd.season_id, wd.week, wd.status
		FROM weekly_drafts wd
		INNER JOIN league_members lm ON wd.member_id = lm.id
		WHERE wd.id = $1 AND lm.user_id = $2
	`, draftID, claims.UserID).Scan(&memberID, &seasonID, &week, &status)
	if err != nil {
		http.Error(w, "draft not found or not authorized", http.StatusForbidden)
		return
	}

	if status == "complete" {
		http.Error(w, "draft is already complete", http.StatusBadRequest)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create draft handler
	handler := &draftMessageHandler{
		server:   s,
		draftID:  draftID,
		seasonID: seasonID,
		memberID: memberID,
		week:     week,
	}

	// Create client
	client := ws.NewClient(s.hub, conn, claims.UserID, memberID, draftID, handler)

	// Register client
	s.hub.Register(client)

	// Start pumps
	go client.WritePump()
	go client.ReadPump()

	// Trigger OnConnect
	handler.OnConnect(client)
}

// draftMessageHandler handles draft WebSocket messages
type draftMessageHandler struct {
	server   *Server
	draftID  uuid.UUID
	seasonID uuid.UUID
	memberID uuid.UUID
	week     int
}

func (h *draftMessageHandler) OnConnect(client *ws.Client) {
	ctx := context.Background()

	// Get or create draft state
	draft, err := h.server.draftManager.GetOrCreateDraft(ctx, h.draftID, h.seasonID, h.memberID, h.week)
	if err != nil {
		client.SendError("failed to load draft")
		return
	}

	// Send current state
	client.SendJSON("draft_state", draft.GetState())

	// If draft is not complete and no current team, draw one
	if !draft.IsComplete() && draft.GetCurrentTeam() == nil {
		h.drawAndSendTeam(client, draft)
	} else if draft.GetCurrentTeam() != nil {
		// Send current team info
		h.sendCurrentTeam(client, draft)
	}
}

func (h *draftMessageHandler) OnDisconnect(client *ws.Client) {
	log.Printf("Client disconnected from draft %s", h.draftID)
}

func (h *draftMessageHandler) HandleMessage(client *ws.Client, messageType string, payload json.RawMessage) {
	ctx := context.Background()

	draft, err := h.server.draftManager.GetOrCreateDraft(ctx, h.draftID, h.seasonID, h.memberID, h.week)
	if err != nil {
		client.SendError("draft not found")
		return
	}

	switch messageType {
	case "draw_team":
		h.handleDrawTeam(client, draft)

	case "make_pick":
		h.handleMakePick(client, draft, payload)

	case "get_state":
		client.SendJSON("draft_state", draft.GetState())

	default:
		client.SendError("unknown message type: " + messageType)
	}
}

func (h *draftMessageHandler) handleDrawTeam(client *ws.Client, draft *game.WeeklyDraftState) {
	if draft.IsComplete() {
		client.SendError("draft is already complete")
		return
	}

	if draft.GetCurrentTeam() != nil {
		// Already have a team drawn, send it
		h.sendCurrentTeam(client, draft)
		return
	}

	h.drawAndSendTeam(client, draft)
}

func (h *draftMessageHandler) drawAndSendTeam(client *ws.Client, draft *game.WeeklyDraftState) {
	team, err := draft.DrawTeam()
	if err != nil {
		client.SendError(err.Error())
		return
	}

	availablePlayers := draft.GetAvailablePlayers()

	client.SendJSON("team_drawn", game.TeamDrawResponse{
		Team: game.TeamInfo{
			Name:   team.Name,
			Abbrev: team.Abbrev,
		},
		AvailablePlayers: availablePlayers,
		CurrentPick:      draft.GetState().CurrentPick,
		TotalPicks:       game.TotalPicks,
	})
}

func (h *draftMessageHandler) sendCurrentTeam(client *ws.Client, draft *game.WeeklyDraftState) {
	team := draft.GetCurrentTeam()
	if team == nil {
		return
	}

	availablePlayers := draft.GetAvailablePlayers()

	client.SendJSON("team_drawn", game.TeamDrawResponse{
		Team: game.TeamInfo{
			Name:   team.Name,
			Abbrev: team.Abbrev,
		},
		AvailablePlayers: availablePlayers,
		CurrentPick:      draft.GetState().CurrentPick,
		TotalPicks:       game.TotalPicks,
	})
}

func (h *draftMessageHandler) handleMakePick(client *ws.Client, draft *game.WeeklyDraftState, payload json.RawMessage) {
	var req struct {
		PlayerID string          `json:"player_id"`
		Position models.Position `json:"position"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		client.SendError("invalid pick payload")
		return
	}

	if req.PlayerID == "" || req.Position == "" {
		client.SendError("player_id and position are required")
		return
	}

	ctx := context.Background()

	pick, err := draft.MakePick(ctx, req.PlayerID, req.Position)
	if err != nil {
		client.SendError(err.Error())
		return
	}

	state := draft.GetState()

	// Send pick confirmation
	client.SendJSON("pick_confirmed", game.PickResponse{
		Pick:        *pick,
		CurrentPick: state.CurrentPick,
		TotalPicks:  game.TotalPicks,
		IsComplete:  state.Status == "complete",
		Roster:      state.Picks,
	})

	// If draft is complete, send final results
	if state.Status == "complete" {
		h.handleDraftComplete(client, draft)
	} else {
		// Draw next team automatically
		h.drawAndSendTeam(client, draft)
	}
}

func (h *draftMessageHandler) handleDraftComplete(client *ws.Client, draft *game.WeeklyDraftState) {
	ctx := context.Background()
	state := draft.GetState()

	// Calculate score (simulated for now)
	score, err := draft.CalculateWeeklyScore(ctx, "ppr")
	if err != nil {
		log.Printf("Failed to calculate score: %v", err)
		score = 0
	}

	// Save weekly result
	_, err = h.server.pool.Exec(ctx, `
		INSERT INTO weekly_results (season_id, member_id, week, total_points, calculated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (season_id, member_id, week) DO UPDATE SET
			total_points = EXCLUDED.total_points,
			calculated_at = NOW()
	`, h.seasonID, h.memberID, h.week, score)
	if err != nil {
		log.Printf("Failed to save weekly result: %v", err)
	}

	// Update standings
	h.updateStandings(ctx, score)

	// Send complete message
	client.SendJSON("draft_complete", map[string]interface{}{
		"draft_id":     h.draftID,
		"week":         h.week,
		"roster":       state.Picks,
		"total_points": score,
		"message":      "Draft complete! Your roster has been submitted.",
	})

	// Remove draft from memory
	h.server.draftManager.RemoveDraft(h.draftID)
}

func (h *draftMessageHandler) updateStandings(ctx context.Context, weekScore float64) {
	// Get current standings
	var currentPoints float64
	var bestWeek float64
	var weeksPlayed int
	var weeklyWins int

	err := h.server.pool.QueryRow(ctx, `
		SELECT total_points, best_week, weeks_played, weekly_wins
		FROM standings
		WHERE season_id = $1 AND member_id = $2
	`, h.seasonID, h.memberID).Scan(&currentPoints, &bestWeek, &weeksPlayed, &weeklyWins)

	if err != nil {
		// Create new standings entry
		currentPoints = 0
		bestWeek = 0
		weeksPlayed = 0
		weeklyWins = 0
	}

	// Update values
	newTotalPoints := currentPoints + weekScore
	newWeeksPlayed := weeksPlayed + 1
	newBestWeek := bestWeek
	if weekScore > bestWeek {
		newBestWeek = weekScore
	}

	// Upsert standings
	_, err = h.server.pool.Exec(ctx, `
		INSERT INTO standings (season_id, member_id, weekly_wins, total_points, best_week, weeks_played, current_rank)
		VALUES ($1, $2, $3, $4, $5, $6, 0)
		ON CONFLICT (season_id, member_id) DO UPDATE SET
			total_points = EXCLUDED.total_points,
			best_week = EXCLUDED.best_week,
			weeks_played = EXCLUDED.weeks_played
	`, h.seasonID, h.memberID, weeklyWins, newTotalPoints, newBestWeek, newWeeksPlayed)
	if err != nil {
		log.Printf("Failed to update standings: %v", err)
	}

	// Recalculate ranks for all members in this season
	h.recalculateRanks(ctx)
}

func (h *draftMessageHandler) recalculateRanks(ctx context.Context) {
	_, err := h.server.pool.Exec(ctx, `
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY total_points DESC) as rank
			FROM standings
			WHERE season_id = $1
		)
		UPDATE standings s
		SET current_rank = r.rank
		FROM ranked r
		WHERE s.id = r.id
	`, h.seasonID)
	if err != nil {
		log.Printf("Failed to recalculate ranks: %v", err)
	}
}

// REST API handlers for drafts

// handleGetDraftRoster returns the roster for a completed draft
func (s *Server) handleGetDraftRoster(w http.ResponseWriter, r *http.Request) {
	draftID, err := uuid.Parse(chi.URLParam(r, "draftID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid draft ID")
		return
	}

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	ctx := r.Context()

	// Verify access and get draft details
	var seasonID, memberID uuid.UUID
	var week int
	var status string
	var completedAt *time.Time
	err = s.pool.QueryRow(ctx, `
		SELECT wd.season_id, wd.member_id, wd.week, wd.status, wd.completed_at
		FROM weekly_drafts wd
		INNER JOIN league_members lm ON wd.member_id = lm.id
		INNER JOIN leagues l ON lm.league_id = l.id
		INNER JOIN league_members lm2 ON l.id = lm2.league_id
		WHERE wd.id = $1 AND lm2.user_id = $2
	`, draftID, userID).Scan(&seasonID, &memberID, &week, &status, &completedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "draft not found")
		return
	}

	// Get picks
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

	// Get score if available
	var totalPoints *float64
	s.pool.QueryRow(ctx, `
		SELECT total_points FROM weekly_results
		WHERE season_id = $1 AND member_id = $2 AND week = $3
	`, seasonID, memberID, week).Scan(&totalPoints)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"draft_id":     draftID,
		"season_id":    seasonID,
		"member_id":    memberID,
		"week":         week,
		"status":       status,
		"completed_at": completedAt,
		"picks":        picks,
		"total_points": totalPoints,
	})
}

// handleGetWeeklyResults returns results for a specific week
func (s *Server) handleGetWeeklyResults(w http.ResponseWriter, r *http.Request) {
	seasonID, err := uuid.Parse(chi.URLParam(r, "seasonID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid season ID")
		return
	}

	week := chi.URLParam(r, "week")

	rows, err := s.pool.Query(r.Context(), `
		SELECT wr.id, wr.member_id, wr.week, wr.total_points, wr.weekly_rank, wr.is_weekly_winner,
		       lm.team_name, u.display_name, u.avatar_url
		FROM weekly_results wr
		INNER JOIN league_members lm ON wr.member_id = lm.id
		INNER JOIN users u ON lm.user_id = u.id
		WHERE wr.season_id = $1 AND wr.week = $2
		ORDER BY wr.total_points DESC
	`, seasonID, week)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch results")
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, memberID uuid.UUID
		var weekNum, weeklyRank int
		var totalPoints float64
		var isWeeklyWinner bool
		var teamName, displayName string
		var avatarURL *string

		if err := rows.Scan(&id, &memberID, &weekNum, &totalPoints, &weeklyRank, &isWeeklyWinner, &teamName, &displayName, &avatarURL); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"id":               id,
			"member_id":        memberID,
			"week":             weekNum,
			"total_points":     totalPoints,
			"weekly_rank":      weeklyRank,
			"is_weekly_winner": isWeeklyWinner,
			"team_name":        teamName,
			"display_name":     displayName,
			"avatar_url":       avatarURL,
		})
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, results)
}
