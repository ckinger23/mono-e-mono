package protocol

import "github.com/ckinger23/mono-e-mono/internal/models"

type MessageType string

const (
	// Server -> Client messages (legacy H2H mode)
	MsgWelcome        MessageType = "welcome"
	MsgWaitingPlayer  MessageType = "waiting_player"
	MsgGameStart      MessageType = "game_start"
	MsgYourTurn       MessageType = "your_turn"
	MsgWaitTurn       MessageType = "wait_turn"
	MsgPickConfirmed  MessageType = "pick_confirmed"
	MsgOpponentPicked MessageType = "opponent_picked"
	MsgGameOver       MessageType = "game_over"
	MsgError          MessageType = "error"
	MsgRosterUpdate   MessageType = "roster_update"

	// Client -> Server messages (legacy H2H mode)
	MsgPickPlayer MessageType = "pick_player"

	// Weekly Draft - Server -> Client messages
	MsgDraftState    MessageType = "draft_state"     // Current draft state
	MsgTeamDrawn     MessageType = "team_drawn"      // Random team drawn for pick
	MsgDraftPick     MessageType = "draft_pick"      // Pick was made
	MsgDraftComplete MessageType = "draft_complete"  // Draft finished

	// Weekly Draft - Client -> Server messages
	MsgDrawTeam MessageType = "draw_team" // Request to draw a team
	MsgMakePick MessageType = "make_pick" // Make a pick
	MsgGetState MessageType = "get_state" // Request current state
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload any         `json:"payload,omitempty"`
}

// Legacy H2H payloads

type WelcomePayload struct {
	PlayerNumber int    `json:"player_number"`
	Message      string `json:"message"`
}

type GameStartPayload struct {
	Message       string `json:"message"`
	YourRoster    string `json:"your_roster"`
	OpponentLabel string `json:"opponent_label"`
}

type YourTurnPayload struct {
	Round            int                 `json:"round"`
	TeamName         string              `json:"team_name"`
	TeamAbbrev       string              `json:"team_abbrev"`
	AvailablePlayers []models.Player     `json:"available_players"`
	OpenPositions    []models.Position   `json:"open_positions"`
	YourRoster       []models.RosterSlot `json:"your_roster"`
}

type WaitTurnPayload struct {
	Round          int    `json:"round"`
	OpponentNumber int    `json:"opponent_number"`
	Message        string `json:"message"`
}

type PickPlayerPayload struct {
	PlayerID string          `json:"player_id"`
	Position models.Position `json:"position"`
}

type PickConfirmedPayload struct {
	PlayerName string          `json:"player_name"`
	Position   models.Position `json:"position"`
	Round      int             `json:"round"`
}

type OpponentPickedPayload struct {
	OpponentNumber int             `json:"opponent_number"`
	PlayerName     string          `json:"player_name"`
	Position       models.Position `json:"position"`
	TeamName       string          `json:"team_name"`
}

type GameOverPayload struct {
	Winner         int                 `json:"winner"`
	YourScore      float64             `json:"your_score"`
	OpponentScore  float64             `json:"opponent_score"`
	YourRoster     []models.RosterSlot `json:"your_roster"`
	OpponentRoster []models.RosterSlot `json:"opponent_roster"`
	Message        string              `json:"message"`
}

type RosterUpdatePayload struct {
	YourRoster     []models.RosterSlot `json:"your_roster"`
	OpponentRoster []models.RosterSlot `json:"opponent_roster"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

// Weekly Draft payloads

type DraftStatePayload struct {
	DraftID         string            `json:"draft_id"`
	SeasonID        string            `json:"season_id"`
	MemberID        string            `json:"member_id"`
	Week            int               `json:"week"`
	Status          string            `json:"status"`
	CurrentPick     int               `json:"current_pick"`
	TotalPicks      int               `json:"total_picks"`
	Picks           []DraftPickInfo   `json:"picks"`
	NeededPositions []models.Position `json:"needed_positions"`
}

type DraftPickInfo struct {
	PickNumber  int    `json:"pick_number"`
	NFLPlayerID string `json:"nfl_player_id"`
	PlayerName  string `json:"player_name"`
	Position    string `json:"position"`
	TeamDrawn   string `json:"team_drawn"`
	PickedAt    string `json:"picked_at"`
}

type TeamDrawnPayload struct {
	Team             TeamInfo        `json:"team"`
	AvailablePlayers []models.Player `json:"available_players"`
	CurrentPick      int             `json:"current_pick"`
	TotalPicks       int             `json:"total_picks"`
}

type TeamInfo struct {
	Name   string `json:"name"`
	Abbrev string `json:"abbrev"`
}

type WeeklyPickConfirmedPayload struct {
	Pick        DraftPickInfo   `json:"pick"`
	CurrentPick int             `json:"current_pick"`
	TotalPicks  int             `json:"total_picks"`
	IsComplete  bool            `json:"is_complete"`
	Roster      []DraftPickInfo `json:"roster"`
}

type DraftCompletePayload struct {
	DraftID     string          `json:"draft_id"`
	Week        int             `json:"week"`
	Roster      []DraftPickInfo `json:"roster"`
	TotalPoints float64         `json:"total_points"`
	Message     string          `json:"message"`
}

type MakePickPayload struct {
	PlayerID string          `json:"player_id"`
	Position models.Position `json:"position"`
}
