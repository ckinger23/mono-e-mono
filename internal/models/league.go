package models

import (
	"time"

	"github.com/google/uuid"
)

// ScoringType represents the fantasy scoring format
type ScoringType string

const (
	ScoringPPR      ScoringType = "ppr"
	ScoringStandard ScoringType = "standard"
	ScoringHalfPPR  ScoringType = "half_ppr"
)

// League represents a fantasy league
type League struct {
	ID             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	InviteCode     string      `json:"invite_code"`
	CommissionerID uuid.UUID   `json:"commissioner_id"`
	MaxMembers     int         `json:"max_members"`
	ScoringType    ScoringType `json:"scoring_type"`
	CreatedAt      time.Time   `json:"created_at"`
}

// LeagueMember represents a user's membership in a league
type LeagueMember struct {
	ID        uuid.UUID `json:"id"`
	LeagueID  uuid.UUID `json:"league_id"`
	UserID    uuid.UUID `json:"user_id"`
	TeamName  string    `json:"team_name"`
	JoinedAt  time.Time `json:"joined_at"`
}

// LeagueMemberWithUser includes user details
type LeagueMemberWithUser struct {
	LeagueMember
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// LeagueResponse is the public representation of a league
type LeagueResponse struct {
	ID             uuid.UUID   `json:"id"`
	Name           string      `json:"name"`
	InviteCode     string      `json:"invite_code,omitempty"`
	CommissionerID uuid.UUID   `json:"commissioner_id"`
	MaxMembers     int         `json:"max_members"`
	ScoringType    ScoringType `json:"scoring_type"`
	MemberCount    int         `json:"member_count,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

// CreateLeagueRequest represents a request to create a league
type CreateLeagueRequest struct {
	Name        string      `json:"name"`
	TeamName    string      `json:"team_name"`
	MaxMembers  int         `json:"max_members,omitempty"`
	ScoringType ScoringType `json:"scoring_type,omitempty"`
}

// JoinLeagueRequest represents a request to join a league
type JoinLeagueRequest struct {
	InviteCode string `json:"invite_code"`
	TeamName   string `json:"team_name"`
}

// UpdateLeagueRequest represents a request to update a league
type UpdateLeagueRequest struct {
	Name        *string      `json:"name,omitempty"`
	MaxMembers  *int         `json:"max_members,omitempty"`
	ScoringType *ScoringType `json:"scoring_type,omitempty"`
}
