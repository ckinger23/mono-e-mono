package nfl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StatsSyncer handles syncing NFL weekly stats from Sleeper API to database
type StatsSyncer struct {
	client *Client
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewStatsSyncer creates a new stats syncer
func NewStatsSyncer(pool *pgxpool.Pool, logger *slog.Logger) *StatsSyncer {
	return &StatsSyncer{
		client: NewClient(),
		pool:   pool,
		logger: logger,
	}
}

// StatsSyncResult contains the results of a stats sync operation
type StatsSyncResult struct {
	Year         int `json:"year"`
	Week         int `json:"week"`
	PlayersSynced int `json:"players_synced"`
	Errors       int `json:"errors"`
}

// SyncWeeklyStats fetches and stores stats for a specific week
func (s *StatsSyncer) SyncWeeklyStats(ctx context.Context, year, week int) (*StatsSyncResult, error) {
	s.logger.Info("starting stats sync from Sleeper API", "year", year, "week", week)

	stats, err := s.client.GetWeeklyStats(ctx, year, week)
	if err != nil {
		return nil, fmt.Errorf("fetching stats from Sleeper: %w", err)
	}

	result := &StatsSyncResult{
		Year: year,
		Week: week,
	}

	batch := &pgx.Batch{}
	batchCount := 0
	const batchSize = 500

	for playerID, stat := range stats {
		// Calculate fantasy points
		ppr := calculatePPRPoints(stat)
		standard := calculateStandardPoints(stat)
		halfPPR := calculateHalfPPRPoints(stat)

		// Convert stats to JSON for storage
		statsJSON, err := json.Marshal(stat)
		if err != nil {
			s.logger.Error("failed to marshal stats", "player_id", playerID, "error", err)
			result.Errors++
			continue
		}

		batch.Queue(
			`INSERT INTO nfl_weekly_stats (player_id, year, week, points_ppr, points_standard, points_half_ppr, stats)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (player_id, year, week) DO UPDATE SET
				points_ppr = EXCLUDED.points_ppr,
				points_standard = EXCLUDED.points_standard,
				points_half_ppr = EXCLUDED.points_half_ppr,
				stats = EXCLUDED.stats`,
			playerID, year, week, ppr, standard, halfPPR, statsJSON,
		)
		batchCount++

		// Execute batch when it reaches size limit
		if batchCount >= batchSize {
			if err := s.executeBatch(ctx, batch); err != nil {
				result.Errors++
				s.logger.Error("batch execution failed", "error", err)
			} else {
				result.PlayersSynced += batchCount
			}
			batch = &pgx.Batch{}
			batchCount = 0
		}
	}

	// Execute remaining batch
	if batchCount > 0 {
		if err := s.executeBatch(ctx, batch); err != nil {
			result.Errors++
			s.logger.Error("final batch execution failed", "error", err)
		} else {
			result.PlayersSynced += batchCount
		}
	}

	s.logger.Info("stats sync complete",
		"year", year,
		"week", week,
		"synced", result.PlayersSynced,
		"errors", result.Errors,
	)

	return result, nil
}

func (s *StatsSyncer) executeBatch(ctx context.Context, batch *pgx.Batch) error {
	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("batch item %d: %w", i, err)
		}
	}

	return nil
}

// DBWeeklyStats represents weekly stats from the database
type DBWeeklyStats struct {
	PlayerID      string          `json:"player_id"`
	Year          int             `json:"year"`
	Week          int             `json:"week"`
	PointsPPR     float64         `json:"points_ppr"`
	PointsStandard float64        `json:"points_standard"`
	PointsHalfPPR float64         `json:"points_half_ppr"`
	Stats         json.RawMessage `json:"stats"`
}

// GetPlayerWeeklyStats returns stats for a specific player in a specific week
func (s *StatsSyncer) GetPlayerWeeklyStats(ctx context.Context, playerID string, year, week int) (*DBWeeklyStats, error) {
	var stats DBWeeklyStats
	err := s.pool.QueryRow(ctx,
		`SELECT player_id, year, week, points_ppr, points_standard, points_half_ppr, stats
		FROM nfl_weekly_stats
		WHERE player_id = $1 AND year = $2 AND week = $3`,
		playerID, year, week,
	).Scan(&stats.PlayerID, &stats.Year, &stats.Week, &stats.PointsPPR, &stats.PointsStandard, &stats.PointsHalfPPR, &stats.Stats)
	if err != nil {
		return nil, fmt.Errorf("querying stats: %w", err)
	}
	return &stats, nil
}

// PlayerWithStats combines player info with weekly stats
type PlayerWithStats struct {
	PlayerID       string  `json:"player_id"`
	Name           string  `json:"name"`
	Position       string  `json:"position"`
	Team           string  `json:"team"`
	PointsPPR      float64 `json:"points_ppr"`
	PointsStandard float64 `json:"points_standard"`
	PointsHalfPPR  float64 `json:"points_half_ppr"`
}

// GetWeeklyLeaders returns top scoring players for a week
func (s *StatsSyncer) GetWeeklyLeaders(ctx context.Context, year, week int, limit int) ([]PlayerWithStats, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT nws.player_id, np.name, np.position, np.team,
			nws.points_ppr, nws.points_standard, nws.points_half_ppr
		FROM nfl_weekly_stats nws
		INNER JOIN nfl_players np ON nws.player_id = np.id
		WHERE nws.year = $1 AND nws.week = $2
		ORDER BY nws.points_ppr DESC
		LIMIT $3`,
		year, week, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying leaders: %w", err)
	}
	defer rows.Close()

	var leaders []PlayerWithStats
	for rows.Next() {
		var p PlayerWithStats
		if err := rows.Scan(&p.PlayerID, &p.Name, &p.Position, &p.Team,
			&p.PointsPPR, &p.PointsStandard, &p.PointsHalfPPR); err != nil {
			return nil, fmt.Errorf("scanning leader: %w", err)
		}
		leaders = append(leaders, p)
	}

	return leaders, rows.Err()
}

// GetCurrentNFLWeek returns the current NFL week from Sleeper
func (s *StatsSyncer) GetCurrentNFLWeek(ctx context.Context) (int, int, error) {
	state, err := s.client.GetNFLState(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("fetching NFL state: %w", err)
	}

	var year int
	if _, err := fmt.Sscanf(state.Season, "%d", &year); err != nil {
		return 0, 0, fmt.Errorf("parsing season year: %w", err)
	}

	return year, state.Week, nil
}

// Helper functions for points calculation
func calculatePPRPoints(s SleeperStats) float64 {
	points := 0.0

	// Passing
	points += s.PassYds * 0.04     // 1 point per 25 yards
	points += s.PassTD * 4         // 4 points per TD
	points += s.PassInt * -2       // -2 points per INT
	points += s.Pass2Pt * 2        // 2 points per 2PT conversion

	// Rushing
	points += s.RushYds * 0.1      // 1 point per 10 yards
	points += s.RushTD * 6         // 6 points per TD
	points += s.Rush2Pt * 2        // 2 points per 2PT conversion

	// Receiving (PPR)
	points += s.Receptions * 1     // 1 point per reception
	points += s.RecYds * 0.1       // 1 point per 10 yards
	points += s.RecTD * 6          // 6 points per TD
	points += s.Rec2Pt * 2         // 2 points per 2PT conversion

	// Fumbles
	points += s.FumblesLost * -2   // -2 points per fumble lost

	// Defense/Special Teams
	points += s.DefTD * 6          // 6 points per defensive TD
	points += s.DefInt * 2         // 2 points per INT
	points += s.DefSack * 1        // 1 point per sack
	points += s.DefFumbRec * 2     // 2 points per fumble recovery
	points += s.DefSafety * 2      // 2 points per safety
	points += s.DefBlkKick * 2     // 2 points per blocked kick

	// Points allowed (defense)
	points += calculatePointsAllowedBonus(s.DefPtsAllowed)

	return points
}

func calculateStandardPoints(s SleeperStats) float64 {
	// Same as PPR but no reception points
	points := calculatePPRPoints(s)
	points -= s.Receptions * 1 // Remove the PPR bonus
	return points
}

func calculateHalfPPRPoints(s SleeperStats) float64 {
	// Same as PPR but only 0.5 per reception
	points := calculatePPRPoints(s)
	points -= s.Receptions * 0.5 // Change from 1 to 0.5
	return points
}

func calculatePointsAllowedBonus(ptsAllowed float64) float64 {
	switch {
	case ptsAllowed == 0:
		return 10
	case ptsAllowed >= 1 && ptsAllowed <= 6:
		return 7
	case ptsAllowed >= 7 && ptsAllowed <= 13:
		return 4
	case ptsAllowed >= 14 && ptsAllowed <= 20:
		return 1
	case ptsAllowed >= 21 && ptsAllowed <= 27:
		return 0
	case ptsAllowed >= 28 && ptsAllowed <= 34:
		return -1
	default:
		return -4
	}
}
