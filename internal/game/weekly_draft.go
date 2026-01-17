package game

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ckinger23/mono-e-mono/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WeeklyDraftState represents the state of a weekly draft
type WeeklyDraftState struct {
	mu sync.RWMutex

	ID          uuid.UUID
	SeasonID    uuid.UUID
	MemberID    uuid.UUID
	Week        int
	Status      string
	CurrentPick int
	Picks       []DraftPickState
	CurrentTeam *models.Team
	UsedTeams   []string

	db       *PlayerDB
	pool     *pgxpool.Pool
	rng      *rand.Rand
}

// DraftPickState represents a single pick in a draft
type DraftPickState struct {
	PickNumber  int           `json:"pick_number"`
	NFLPlayerID string        `json:"nfl_player_id"`
	PlayerName  string        `json:"player_name"`
	Position    string        `json:"position"`
	TeamDrawn   string        `json:"team_drawn"`
	PickedAt    time.Time     `json:"picked_at"`
}

// RosterRequirements defines the roster slots needed
var RosterRequirements = []models.Position{
	models.PositionQB,
	models.PositionRB,
	models.PositionRB,
	models.PositionWR,
	models.PositionWR,
	models.PositionTE,
	models.PositionDEF,
}

const TotalPicks = 7

// WeeklyDraftManager manages all active drafts
type WeeklyDraftManager struct {
	mu     sync.RWMutex
	drafts map[uuid.UUID]*WeeklyDraftState
	db     *PlayerDB
	pool   *pgxpool.Pool
}

// NewWeeklyDraftManager creates a new draft manager
func NewWeeklyDraftManager(db *PlayerDB, pool *pgxpool.Pool) *WeeklyDraftManager {
	return &WeeklyDraftManager{
		drafts: make(map[uuid.UUID]*WeeklyDraftState),
		db:     db,
		pool:   pool,
	}
}

// GetOrCreateDraft gets an existing draft or creates a new one
func (m *WeeklyDraftManager) GetOrCreateDraft(ctx context.Context, draftID, seasonID, memberID uuid.UUID, week int) (*WeeklyDraftState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if draft is already in memory
	if draft, ok := m.drafts[draftID]; ok {
		return draft, nil
	}

	// Load draft from database
	draft, err := m.loadDraftFromDB(ctx, draftID)
	if err != nil {
		return nil, err
	}

	m.drafts[draftID] = draft
	return draft, nil
}

func (m *WeeklyDraftManager) loadDraftFromDB(ctx context.Context, draftID uuid.UUID) (*WeeklyDraftState, error) {
	// Get draft info
	var seasonID, memberID uuid.UUID
	var week int
	var status string
	err := m.pool.QueryRow(ctx, `
		SELECT season_id, member_id, week, status
		FROM weekly_drafts WHERE id = $1
	`, draftID).Scan(&seasonID, &memberID, &week, &status)
	if err != nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	draft := &WeeklyDraftState{
		ID:          draftID,
		SeasonID:    seasonID,
		MemberID:    memberID,
		Week:        week,
		Status:      status,
		CurrentPick: 1,
		Picks:       []DraftPickState{},
		UsedTeams:   []string{},
		db:          m.db,
		pool:        m.pool,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Load existing picks
	rows, err := m.pool.Query(ctx, `
		SELECT pick_number, nfl_player_id, player_name, position, team_drawn, picked_at
		FROM draft_picks WHERE weekly_draft_id = $1 ORDER BY pick_number
	`, draftID)
	if err != nil {
		return nil, fmt.Errorf("failed to load picks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pick DraftPickState
		if err := rows.Scan(&pick.PickNumber, &pick.NFLPlayerID, &pick.PlayerName, &pick.Position, &pick.TeamDrawn, &pick.PickedAt); err != nil {
			continue
		}
		draft.Picks = append(draft.Picks, pick)
		draft.UsedTeams = append(draft.UsedTeams, pick.TeamDrawn)
	}

	draft.CurrentPick = len(draft.Picks) + 1

	return draft, nil
}

// RemoveDraft removes a draft from memory
func (m *WeeklyDraftManager) RemoveDraft(draftID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.drafts, draftID)
}

// DrawTeam draws a random team for the current pick
func (d *WeeklyDraftState) DrawTeam() (*models.Team, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Status == "complete" {
		return nil, fmt.Errorf("draft is already complete")
	}

	if d.CurrentPick > TotalPicks {
		return nil, fmt.Errorf("all picks have been made")
	}

	// Get all team names
	teamNames := d.db.GetTeamNames()

	// Shuffle teams
	d.rng.Shuffle(len(teamNames), func(i, j int) {
		teamNames[i], teamNames[j] = teamNames[j], teamNames[i]
	})

	// Pick a random team
	idx := d.rng.Intn(len(teamNames))
	team := d.db.GetTeam(teamNames[idx])
	d.CurrentTeam = team

	return team, nil
}

// GetCurrentTeam returns the current team for the draft
func (d *WeeklyDraftState) GetCurrentTeam() *models.Team {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.CurrentTeam
}

// GetAvailablePlayers returns available players from the current team
func (d *WeeklyDraftState) GetAvailablePlayers() []models.Player {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.CurrentTeam == nil {
		return nil
	}

	// Get positions still needed
	neededPositions := d.getNeededPositions()

	var available []models.Player
	for _, player := range d.CurrentTeam.Players {
		// Check if this position is still needed
		if _, needed := neededPositions[player.Position]; needed {
			available = append(available, player)
		}
	}

	return available
}

func (d *WeeklyDraftState) getNeededPositions() map[models.Position]int {
	needed := make(map[models.Position]int)
	for _, pos := range RosterRequirements {
		needed[pos]++
	}

	// Subtract positions already filled
	for _, pick := range d.Picks {
		pos := models.Position(pick.Position)
		if count, ok := needed[pos]; ok && count > 0 {
			needed[pos]--
			if needed[pos] == 0 {
				delete(needed, pos)
			}
		}
	}

	return needed
}

// GetNeededPositions returns positions still needed (exported for API use)
func (d *WeeklyDraftState) GetNeededPositions() []models.Position {
	d.mu.RLock()
	defer d.mu.RUnlock()

	neededMap := d.getNeededPositions()
	var positions []models.Position
	for pos, count := range neededMap {
		for i := 0; i < count; i++ {
			positions = append(positions, pos)
		}
	}
	return positions
}

// MakePick makes a pick in the draft
func (d *WeeklyDraftState) MakePick(ctx context.Context, playerID string, position models.Position) (*DraftPickState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Status == "complete" {
		return nil, fmt.Errorf("draft is already complete")
	}

	if d.CurrentTeam == nil {
		return nil, fmt.Errorf("no team has been drawn")
	}

	// Validate player exists on current team
	var selectedPlayer *models.Player
	for _, p := range d.CurrentTeam.Players {
		if p.ID == playerID {
			selectedPlayer = &p
			break
		}
	}

	if selectedPlayer == nil {
		return nil, fmt.Errorf("player not found on current team")
	}

	// Validate position matches
	if selectedPlayer.Position != position {
		return nil, fmt.Errorf("player position mismatch: expected %s, got %s", position, selectedPlayer.Position)
	}

	// Validate position is still needed
	neededPositions := d.getNeededPositions()
	if _, needed := neededPositions[position]; !needed {
		return nil, fmt.Errorf("position %s is already filled", position)
	}

	// Create pick
	pick := DraftPickState{
		PickNumber:  d.CurrentPick,
		NFLPlayerID: playerID,
		PlayerName:  selectedPlayer.Name,
		Position:    string(position),
		TeamDrawn:   d.CurrentTeam.Name,
		PickedAt:    time.Now(),
	}

	// Save to database
	_, err := d.pool.Exec(ctx, `
		INSERT INTO draft_picks (weekly_draft_id, pick_number, nfl_player_id, player_name, position, team_drawn, picked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, d.ID, pick.PickNumber, pick.NFLPlayerID, pick.PlayerName, pick.Position, pick.TeamDrawn, pick.PickedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save pick: %w", err)
	}

	// Update state
	d.Picks = append(d.Picks, pick)
	d.UsedTeams = append(d.UsedTeams, d.CurrentTeam.Name)
	d.CurrentPick++
	d.CurrentTeam = nil

	// Check if draft is complete
	if d.CurrentPick > TotalPicks {
		d.Status = "complete"
		_, err = d.pool.Exec(ctx, `
			UPDATE weekly_drafts SET status = 'complete', completed_at = NOW()
			WHERE id = $1
		`, d.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to complete draft: %w", err)
		}
	}

	return &pick, nil
}

// GetState returns the current draft state for the client
func (d *WeeklyDraftState) GetState() *DraftStateResponse {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return &DraftStateResponse{
		DraftID:         d.ID,
		SeasonID:        d.SeasonID,
		MemberID:        d.MemberID,
		Week:            d.Week,
		Status:          d.Status,
		CurrentPick:     d.CurrentPick,
		TotalPicks:      TotalPicks,
		Picks:           d.Picks,
		NeededPositions: d.GetNeededPositions(),
	}
}

// IsComplete returns whether the draft is complete
func (d *WeeklyDraftState) IsComplete() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Status == "complete"
}

// DraftStateResponse is the API response for draft state
type DraftStateResponse struct {
	DraftID         uuid.UUID         `json:"draft_id"`
	SeasonID        uuid.UUID         `json:"season_id"`
	MemberID        uuid.UUID         `json:"member_id"`
	Week            int               `json:"week"`
	Status          string            `json:"status"`
	CurrentPick     int               `json:"current_pick"`
	TotalPicks      int               `json:"total_picks"`
	Picks           []DraftPickState  `json:"picks"`
	NeededPositions []models.Position `json:"needed_positions"`
}

// TeamDrawResponse is the API response for a team draw
type TeamDrawResponse struct {
	Team             TeamInfo        `json:"team"`
	AvailablePlayers []models.Player `json:"available_players"`
	CurrentPick      int             `json:"current_pick"`
	TotalPicks       int             `json:"total_picks"`
}

// TeamInfo is basic team information
type TeamInfo struct {
	Name   string `json:"name"`
	Abbrev string `json:"abbrev"`
}

// PickResponse is the API response for a pick
type PickResponse struct {
	Pick        DraftPickState    `json:"pick"`
	CurrentPick int               `json:"current_pick"`
	TotalPicks  int               `json:"total_picks"`
	IsComplete  bool              `json:"is_complete"`
	Roster      []DraftPickState  `json:"roster"`
}

// CalculateWeeklyScore calculates the score for a completed draft
func (d *WeeklyDraftState) CalculateWeeklyScore(ctx context.Context, scoringType string) (float64, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.Status != "complete" {
		return 0, fmt.Errorf("draft is not complete")
	}

	// For now, we'll simulate scores since we don't have real NFL data yet
	// In Phase 4, this will use real Sleeper API data
	config := DefaultScoringConfig()
	var totalScore float64

	for _, pick := range d.Picks {
		player := d.db.GetPlayer(pick.NFLPlayerID)
		if player == nil {
			continue
		}

		stats := SimulateWeekStats(player)
		score := stats.CalculateScore(config, player.Position)
		totalScore += score
	}

	return totalScore, nil
}

// ToJSON converts the draft state to JSON for storage/caching
func (d *WeeklyDraftState) ToJSON() ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return json.Marshal(d.GetState())
}
