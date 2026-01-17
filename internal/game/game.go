package game

import (
	"fmt"
	"sync"

	"github.com/ckinger23/mono-e-mono/internal/models"
)

type GameState int

const (
	StateWaitingForPlayers GameState = iota
	StateInProgress
	StateComplete
)

type PlayerState struct {
	ID     int
	Name   string
	Roster *models.Roster
	Ready  bool
}

type Game struct {
	mu          sync.RWMutex
	ID          string
	State       GameState
	Players     [2]*PlayerState
	PlayerCount int
	CurrentTurn int
	Round       int
	Draft       *Draft
	CurrentTeam *models.Team
}

func NewGame(id string, db *PlayerDB) *Game {
	return &Game{
		ID:          id,
		State:       StateWaitingForPlayers,
		Players:     [2]*PlayerState{},
		PlayerCount: 0,
		CurrentTurn: 0,
		Round:       1,
		Draft:       NewDraft(db),
	}
}

func (g *Game) AddPlayer(name string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.PlayerCount >= 2 {
		return -1, fmt.Errorf("game is full")
	}

	playerID := g.PlayerCount
	g.Players[playerID] = &PlayerState{
		ID:     playerID,
		Name:   name,
		Roster: models.NewRoster(),
		Ready:  false,
	}
	g.PlayerCount++

	if g.PlayerCount == 2 {
		g.State = StateInProgress
		g.drawNextTeam()
	}

	return playerID, nil
}

func (g *Game) GetPlayer(playerID int) *PlayerState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if playerID < 0 || playerID >= 2 {
		return nil
	}
	return g.Players[playerID]
}

func (g *Game) IsPlayerTurn(playerID int) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.CurrentTurn == playerID
}

func (g *Game) GetState() GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.State
}

func (g *Game) GetRound() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Round
}

func (g *Game) GetCurrentTeam() *models.Team {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.CurrentTeam
}

func (g *Game) GetCurrentTurn() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.CurrentTurn
}

func (g *Game) drawNextTeam() {
	g.CurrentTeam = g.Draft.DrawRandomTeam()
}

func (g *Game) MakePick(playerID int, playerPickID string, position models.Position) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StateInProgress {
		return fmt.Errorf("game is not in progress")
	}

	if g.CurrentTurn != playerID {
		return fmt.Errorf("not your turn")
	}

	player := g.Players[playerID]
	if player == nil {
		return fmt.Errorf("player not found")
	}

	if !player.Roster.HasOpenSlot(position) {
		return fmt.Errorf("no open slot for position %s", position)
	}

	draftedPlayer, err := g.Draft.DraftPlayer(playerPickID)
	if err != nil {
		return err
	}

	if draftedPlayer.Position != position {
		return fmt.Errorf("player %s is a %s, not a %s", draftedPlayer.Name, draftedPlayer.Position, position)
	}

	if err := player.Roster.FillSlot(position, draftedPlayer); err != nil {
		return err
	}

	g.advanceTurn()
	return nil
}

func (g *Game) advanceTurn() {
	g.CurrentTurn = (g.CurrentTurn + 1) % 2

	// If we've gone back to player 0, increment the round
	if g.CurrentTurn == 0 {
		g.Round++
	}

	// Check if game is complete (7 slots per player = 7 rounds)
	if g.Round > 7 {
		g.State = StateComplete
		return
	}

	// Draw a new team for the next player's turn
	g.drawNextTeam()
}

func (g *Game) IsComplete() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.State == StateComplete
}

func (g *Game) GetAvailablePlayers() []models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Draft.GetAvailablePlayers(g.CurrentTeam)
}

func (g *Game) GetAvailablePlayersForRoster(playerID int) []models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	available := g.Draft.GetAvailablePlayers(g.CurrentTeam)
	roster := g.Players[playerID].Roster
	return g.Draft.FilterAvailableForRoster(available, roster)
}

func (g *Game) IsFull() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.PlayerCount >= 2
}
