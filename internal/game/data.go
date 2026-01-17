package game

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/ckinger23/mono-e-mono/internal/models"
)

//go:embed players.json
var playersFS embed.FS

type PlayersData struct {
	Teams []models.Team `json:"teams"`
}

type PlayerDB struct {
	Teams       []models.Team
	TeamsByName map[string]*models.Team
	AllPlayers  map[string]*models.Player
}

func LoadPlayerDB() (*PlayerDB, error) {
	data, err := playersFS.ReadFile("players.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded players.json: %w", err)
	}

	var playersData PlayersData
	if err := json.Unmarshal(data, &playersData); err != nil {
		return nil, fmt.Errorf("failed to parse players.json: %w", err)
	}

	db := &PlayerDB{
		Teams:       playersData.Teams,
		TeamsByName: make(map[string]*models.Team),
		AllPlayers:  make(map[string]*models.Player),
	}

	for i := range db.Teams {
		team := &db.Teams[i]
		db.TeamsByName[team.Name] = team
		db.TeamsByName[team.Abbrev] = team

		for j := range team.Players {
			player := &team.Players[j]
			db.AllPlayers[player.ID] = player
		}
	}

	return db, nil
}

func (db *PlayerDB) GetTeam(nameOrAbbrev string) *models.Team {
	return db.TeamsByName[nameOrAbbrev]
}

func (db *PlayerDB) GetPlayer(playerID string) *models.Player {
	return db.AllPlayers[playerID]
}

func (db *PlayerDB) GetTeamNames() []string {
	names := make([]string, len(db.Teams))
	for i, team := range db.Teams {
		names[i] = team.Name
	}
	return names
}
