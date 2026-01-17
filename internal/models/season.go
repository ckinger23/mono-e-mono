package models

import (
	"time"

	"github.com/google/uuid"
)

// SeasonStatus represents the status of a season
type SeasonStatus string

const (
	SeasonStatusActive   SeasonStatus = "active"
	SeasonStatusComplete SeasonStatus = "complete"
)

// Season represents a fantasy season within a league
type Season struct {
	ID               uuid.UUID    `json:"id"`
	LeagueID         uuid.UUID    `json:"league_id"`
	Year             int          `json:"year"`
	CurrentWeek      int          `json:"current_week"`
	Status           SeasonStatus `json:"status"`
	ChampionMemberID *uuid.UUID   `json:"champion_member_id,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
}

// Standing represents a member's standing in a season
type Standing struct {
	ID          uuid.UUID `json:"id"`
	SeasonID    uuid.UUID `json:"season_id"`
	MemberID    uuid.UUID `json:"member_id"`
	WeeklyWins  int       `json:"weekly_wins"`
	TotalPoints float64   `json:"total_points"`
	BestWeek    float64   `json:"best_week"`
	WeeksPlayed int       `json:"weeks_played"`
	CurrentRank int       `json:"current_rank"`
}

// StandingWithMember includes member details
type StandingWithMember struct {
	Standing
	TeamName    string  `json:"team_name"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// DraftStatus represents the status of a weekly draft
type DraftStatus string

const (
	DraftStatusPending    DraftStatus = "pending"
	DraftStatusInProgress DraftStatus = "in_progress"
	DraftStatusComplete   DraftStatus = "complete"
)

// WeeklyDraft represents a user's draft for a specific week
type WeeklyDraft struct {
	ID          uuid.UUID   `json:"id"`
	SeasonID    uuid.UUID   `json:"season_id"`
	MemberID    uuid.UUID   `json:"member_id"`
	Week        int         `json:"week"`
	Status      DraftStatus `json:"status"`
	StartedAt   *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
}

// DraftPick represents a single pick in a weekly draft
type DraftPick struct {
	ID            uuid.UUID `json:"id"`
	WeeklyDraftID uuid.UUID `json:"weekly_draft_id"`
	PickNumber    int       `json:"pick_number"`
	NFLPlayerID   string    `json:"nfl_player_id"`
	PlayerName    string    `json:"player_name"`
	Position      string    `json:"position"`
	TeamDrawn     string    `json:"team_drawn"`
	PickedAt      time.Time `json:"picked_at"`
}

// WeeklyResult represents a member's score for a week
type WeeklyResult struct {
	ID             uuid.UUID  `json:"id"`
	SeasonID       uuid.UUID  `json:"season_id"`
	MemberID       uuid.UUID  `json:"member_id"`
	Week           int        `json:"week"`
	TotalPoints    float64    `json:"total_points"`
	WeeklyRank     int        `json:"weekly_rank"`
	IsWeeklyWinner bool       `json:"is_weekly_winner"`
	CalculatedAt   *time.Time `json:"calculated_at,omitempty"`
}

// WeeklyResultWithMember includes member details
type WeeklyResultWithMember struct {
	WeeklyResult
	TeamName    string  `json:"team_name"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// CreateSeasonRequest represents a request to create a season
type CreateSeasonRequest struct {
	Year int `json:"year"`
}
