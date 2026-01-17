package models

type Team struct {
	Name    string   `json:"name"`
	Abbrev  string   `json:"abbrev"`
	Players []Player `json:"players"`
}

func (t *Team) GetPlayersByPosition(pos Position) []Player {
	var players []Player
	for _, p := range t.Players {
		if p.Position == pos {
			players = append(players, p)
		}
	}
	return players
}

func (t *Team) GetPlayer(playerID string) *Player {
	for _, p := range t.Players {
		if p.ID == playerID {
			return &p
		}
	}
	return nil
}
