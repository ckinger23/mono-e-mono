package game

import (
	"fmt"
	"math/rand"

	"github.com/ckinger23/mono-e-mono/internal/models"
)

type Draft struct {
	DB            *PlayerDB
	DraftedPlayers map[string]bool // player ID -> drafted
	UsedTeams      []string        // teams that have been drawn
}

func NewDraft(db *PlayerDB) *Draft {
	return &Draft{
		DB:            db,
		DraftedPlayers: make(map[string]bool),
		UsedTeams:      []string{},
	}
}

func (d *Draft) DrawRandomTeam() *models.Team {
	teamNames := d.DB.GetTeamNames()

	// Shuffle the team names
	for i := len(teamNames) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		teamNames[i], teamNames[j] = teamNames[j], teamNames[i]
	}

	// Pick a random team (teams can be repeated since players get removed)
	idx := rand.Intn(len(teamNames))
	team := d.DB.GetTeam(teamNames[idx])
	d.UsedTeams = append(d.UsedTeams, team.Name)
	return team
}

func (d *Draft) GetAvailablePlayers(team *models.Team) []models.Player {
	var available []models.Player
	for _, player := range team.Players {
		if !d.DraftedPlayers[player.ID] {
			available = append(available, player)
		}
	}
	return available
}

func (d *Draft) GetAvailablePlayersByPosition(team *models.Team, pos models.Position) []models.Player {
	var available []models.Player
	for _, player := range team.Players {
		if !d.DraftedPlayers[player.ID] && player.Position == pos {
			available = append(available, player)
		}
	}
	return available
}

func (d *Draft) DraftPlayer(playerID string) (*models.Player, error) {
	player := d.DB.GetPlayer(playerID)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	if d.DraftedPlayers[playerID] {
		return nil, fmt.Errorf("player already drafted: %s", player.Name)
	}

	d.DraftedPlayers[playerID] = true
	return player, nil
}

func (d *Draft) IsPlayerAvailable(playerID string) bool {
	return !d.DraftedPlayers[playerID]
}

func (d *Draft) FilterAvailableForRoster(players []models.Player, roster *models.Roster) []models.Player {
	var filtered []models.Player
	for _, player := range players {
		if roster.HasOpenSlot(player.Position) {
			filtered = append(filtered, player)
		}
	}
	return filtered
}
