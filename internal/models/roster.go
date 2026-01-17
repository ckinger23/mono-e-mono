package models

import "fmt"

type RosterSlot struct {
	Position Position `json:"position"`
	Player   *Player  `json:"player,omitempty"`
}

type Roster struct {
	Slots []RosterSlot `json:"slots"`
}

func NewRoster() *Roster {
	return &Roster{
		Slots: []RosterSlot{
			{Position: PositionQB, Player: nil},
			{Position: PositionRB, Player: nil},
			{Position: PositionRB, Player: nil},
			{Position: PositionWR, Player: nil},
			{Position: PositionWR, Player: nil},
			{Position: PositionTE, Player: nil},
			{Position: PositionDEF, Player: nil},
		},
	}
}

func (r *Roster) GetOpenSlots() []RosterSlot {
	var open []RosterSlot
	for _, slot := range r.Slots {
		if slot.Player == nil {
			open = append(open, slot)
		}
	}
	return open
}

func (r *Roster) GetOpenPositions() []Position {
	positionCount := make(map[Position]int)
	for _, slot := range r.Slots {
		if slot.Player == nil {
			positionCount[slot.Position]++
		}
	}

	var positions []Position
	for pos, count := range positionCount {
		for i := 0; i < count; i++ {
			positions = append(positions, pos)
		}
	}
	return positions
}

func (r *Roster) HasOpenSlot(pos Position) bool {
	for _, slot := range r.Slots {
		if slot.Position == pos && slot.Player == nil {
			return true
		}
	}
	return false
}

func (r *Roster) FillSlot(pos Position, player *Player) error {
	for i, slot := range r.Slots {
		if slot.Position == pos && slot.Player == nil {
			r.Slots[i].Player = player
			return nil
		}
	}
	return fmt.Errorf("no open slot for position %s", pos)
}

func (r *Roster) IsFull() bool {
	for _, slot := range r.Slots {
		if slot.Player == nil {
			return false
		}
	}
	return true
}

func (r *Roster) GetFilledSlots() []RosterSlot {
	var filled []RosterSlot
	for _, slot := range r.Slots {
		if slot.Player != nil {
			filled = append(filled, slot)
		}
	}
	return filled
}
