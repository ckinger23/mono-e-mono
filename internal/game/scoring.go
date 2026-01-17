package game

import (
	"math/rand"

	"github.com/ckinger23/mono-e-mono/internal/models"
)

// ScoringConfig defines fantasy scoring rules (PPR-style)
type ScoringConfig struct {
	PassingYardsPerPoint float64
	PassingTD            float64
	Interception         float64
	RushingYardsPerPoint float64
	RushingTD            float64
	Reception            float64 // PPR
	ReceivingYardsPerPoint float64
	ReceivingTD          float64
	Fumble               float64
	DefenseTD            float64
	DefenseSack          float64
	DefenseInterception  float64
	DefensePointsAllowed map[int]float64 // points allowed -> fantasy points
}

func DefaultScoringConfig() *ScoringConfig {
	return &ScoringConfig{
		PassingYardsPerPoint:   25,  // 1 point per 25 yards
		PassingTD:              4,
		Interception:           -2,
		RushingYardsPerPoint:   10,  // 1 point per 10 yards
		RushingTD:              6,
		Reception:              1,   // PPR
		ReceivingYardsPerPoint: 10,  // 1 point per 10 yards
		ReceivingTD:            6,
		Fumble:                 -2,
		DefenseTD:              6,
		DefenseSack:            1,
		DefenseInterception:    2,
		DefensePointsAllowed: map[int]float64{
			0:  10,
			6:  7,
			13: 4,
			20: 1,
			27: 0,
			34: -1,
			99: -4,
		},
	}
}

// PlayerWeekStats represents a player's stats for a week
type PlayerWeekStats struct {
	PassingYards    int
	PassingTDs      int
	Interceptions   int
	RushingYards    int
	RushingTDs      int
	Receptions      int
	ReceivingYards  int
	ReceivingTDs    int
	Fumbles         int
	// Defense-specific
	Sacks           int
	DefenseINTs     int
	DefenseTDs      int
	PointsAllowed   int
}

func (s *PlayerWeekStats) CalculateScore(config *ScoringConfig, position models.Position) float64 {
	var score float64

	// Passing (mainly QB)
	score += float64(s.PassingYards) / config.PassingYardsPerPoint
	score += float64(s.PassingTDs) * config.PassingTD
	score += float64(s.Interceptions) * config.Interception

	// Rushing (RB, some QB)
	score += float64(s.RushingYards) / config.RushingYardsPerPoint
	score += float64(s.RushingTDs) * config.RushingTD

	// Receiving (WR, TE, RB)
	score += float64(s.Receptions) * config.Reception
	score += float64(s.ReceivingYards) / config.ReceivingYardsPerPoint
	score += float64(s.ReceivingTDs) * config.ReceivingTD

	// Fumbles
	score += float64(s.Fumbles) * config.Fumble

	// Defense
	if position == models.PositionDEF {
		score += float64(s.Sacks) * config.DefenseSack
		score += float64(s.DefenseINTs) * config.DefenseInterception
		score += float64(s.DefenseTDs) * config.DefenseTD

		// Points allowed scoring
		for threshold, points := range config.DefensePointsAllowed {
			if s.PointsAllowed <= threshold {
				score += points
				break
			}
		}
	}

	return score
}

// SimulateWeekStats generates random but realistic stats for a player
// In a real app, this would fetch from an API
func SimulateWeekStats(player *models.Player) *PlayerWeekStats {
	stats := &PlayerWeekStats{}

	switch player.Position {
	case models.PositionQB:
		stats.PassingYards = rand.Intn(200) + 150  // 150-350 yards
		stats.PassingTDs = rand.Intn(4)            // 0-3 TDs
		stats.Interceptions = rand.Intn(3)         // 0-2 INTs
		stats.RushingYards = rand.Intn(40)         // 0-40 yards
		if rand.Float64() < 0.15 {
			stats.RushingTDs = 1
		}

	case models.PositionRB:
		stats.RushingYards = rand.Intn(100) + 20   // 20-120 yards
		stats.RushingTDs = rand.Intn(2)            // 0-1 TDs
		stats.Receptions = rand.Intn(6)            // 0-5 receptions
		stats.ReceivingYards = rand.Intn(50)       // 0-50 yards
		if rand.Float64() < 0.1 {
			stats.ReceivingTDs = 1
		}

	case models.PositionWR:
		stats.Receptions = rand.Intn(8) + 2        // 2-10 receptions
		stats.ReceivingYards = rand.Intn(100) + 30 // 30-130 yards
		stats.ReceivingTDs = rand.Intn(2)          // 0-1 TDs

	case models.PositionTE:
		stats.Receptions = rand.Intn(6) + 1        // 1-7 receptions
		stats.ReceivingYards = rand.Intn(70) + 10  // 10-80 yards
		if rand.Float64() < 0.25 {
			stats.ReceivingTDs = 1
		}

	case models.PositionDEF:
		stats.Sacks = rand.Intn(5)                 // 0-4 sacks
		stats.DefenseINTs = rand.Intn(3)           // 0-2 INTs
		if rand.Float64() < 0.1 {
			stats.DefenseTDs = 1
		}
		stats.PointsAllowed = rand.Intn(35)        // 0-35 points allowed
	}

	// Small chance of fumble for skill positions
	if player.Position != models.PositionDEF && rand.Float64() < 0.05 {
		stats.Fumbles = 1
	}

	return stats
}

// CalculateRosterScore calculates total score for a roster
func CalculateRosterScore(roster *models.Roster, config *ScoringConfig) (float64, map[string]float64) {
	var totalScore float64
	playerScores := make(map[string]float64)

	for _, slot := range roster.Slots {
		if slot.Player != nil {
			stats := SimulateWeekStats(slot.Player)
			score := stats.CalculateScore(config, slot.Player.Position)
			playerScores[slot.Player.ID] = score
			totalScore += score
		}
	}

	return totalScore, playerScores
}
