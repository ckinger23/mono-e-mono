package models

type Position string

const (
	PositionQB  Position = "QB"
	PositionRB  Position = "RB"
	PositionWR  Position = "WR"
	PositionTE  Position = "TE"
	PositionDEF Position = "DEF"
)

type Player struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position Position `json:"position"`
	Team     string   `json:"team"`
}

func (p Position) IsValid() bool {
	switch p {
	case PositionQB, PositionRB, PositionWR, PositionTE, PositionDEF:
		return true
	}
	return false
}

func AllPositions() []Position {
	return []Position{PositionQB, PositionRB, PositionWR, PositionTE, PositionDEF}
}
