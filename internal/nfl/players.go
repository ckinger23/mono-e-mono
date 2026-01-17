package nfl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlayerSyncer handles syncing NFL players from Sleeper API to database
type PlayerSyncer struct {
	client *Client
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPlayerSyncer creates a new player syncer
func NewPlayerSyncer(pool *pgxpool.Pool, logger *slog.Logger) *PlayerSyncer {
	return &PlayerSyncer{
		client: NewClient(),
		pool:   pool,
		logger: logger,
	}
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	TotalPlayers   int `json:"total_players"`
	PlayersCreated int `json:"players_created"`
	PlayersUpdated int `json:"players_updated"`
	PlayersSkipped int `json:"players_skipped"`
	Errors         int `json:"errors"`
}

// SyncAllPlayers fetches all players from Sleeper and syncs to database
func (s *PlayerSyncer) SyncAllPlayers(ctx context.Context) (*SyncResult, error) {
	s.logger.Info("starting player sync from Sleeper API")

	players, err := s.client.GetAllPlayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching players from Sleeper: %w", err)
	}

	result := &SyncResult{
		TotalPlayers: len(players),
	}

	// Filter to only fantasy-relevant players
	relevantPositions := map[string]bool{
		"QB":  true,
		"RB":  true,
		"WR":  true,
		"TE":  true,
		"K":   true,
		"DEF": true,
	}

	batch := &pgx.Batch{}
	batchCount := 0
	const batchSize = 500

	for id, player := range players {
		// Skip non-fantasy positions and inactive players
		if !relevantPositions[player.Position] {
			result.PlayersSkipped++
			continue
		}

		// Skip players without a team (free agents)
		if player.Team == "" {
			result.PlayersSkipped++
			continue
		}

		name := player.FullName
		if name == "" {
			name = fmt.Sprintf("%s %s", player.FirstName, player.LastName)
		}

		status := player.Status
		if status == "" {
			status = "active"
		}

		batch.Queue(
			`INSERT INTO nfl_players (id, name, position, team, status, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				position = EXCLUDED.position,
				team = EXCLUDED.team,
				status = EXCLUDED.status,
				updated_at = NOW()`,
			id, name, player.Position, player.Team, status,
		)
		batchCount++

		// Execute batch when it reaches size limit
		if batchCount >= batchSize {
			if err := s.executeBatch(ctx, batch); err != nil {
				result.Errors++
				s.logger.Error("batch execution failed", "error", err)
			} else {
				result.PlayersUpdated += batchCount
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
			result.PlayersUpdated += batchCount
		}
	}

	s.logger.Info("player sync complete",
		"total", result.TotalPlayers,
		"synced", result.PlayersUpdated,
		"skipped", result.PlayersSkipped,
		"errors", result.Errors,
	)

	return result, nil
}

func (s *PlayerSyncer) executeBatch(ctx context.Context, batch *pgx.Batch) error {
	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("batch item %d: %w", i, err)
		}
	}

	return nil
}

// DBPlayer represents a player from the database
type DBPlayer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Team     string `json:"team"`
	Status   string `json:"status"`
}

// GetPlayersByTeam returns all players for a given NFL team
func (s *PlayerSyncer) GetPlayersByTeam(ctx context.Context, team string) ([]DBPlayer, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, position, team, status
		FROM nfl_players
		WHERE team = $1
		ORDER BY position, name`,
		team,
	)
	if err != nil {
		return nil, fmt.Errorf("querying players: %w", err)
	}
	defer rows.Close()

	var players []DBPlayer
	for rows.Next() {
		var p DBPlayer
		if err := rows.Scan(&p.ID, &p.Name, &p.Position, &p.Team, &p.Status); err != nil {
			return nil, fmt.Errorf("scanning player: %w", err)
		}
		players = append(players, p)
	}

	return players, rows.Err()
}

// GetPlayersByPosition returns all players at a given position
func (s *PlayerSyncer) GetPlayersByPosition(ctx context.Context, position string) ([]DBPlayer, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, position, team, status
		FROM nfl_players
		WHERE position = $1
		ORDER BY team, name`,
		position,
	)
	if err != nil {
		return nil, fmt.Errorf("querying players: %w", err)
	}
	defer rows.Close()

	var players []DBPlayer
	for rows.Next() {
		var p DBPlayer
		if err := rows.Scan(&p.ID, &p.Name, &p.Position, &p.Team, &p.Status); err != nil {
			return nil, fmt.Errorf("scanning player: %w", err)
		}
		players = append(players, p)
	}

	return players, rows.Err()
}

// GetPlayer returns a single player by ID
func (s *PlayerSyncer) GetPlayer(ctx context.Context, id string) (*DBPlayer, error) {
	var p DBPlayer
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, position, team, status FROM nfl_players WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Position, &p.Team, &p.Status)
	if err != nil {
		return nil, fmt.Errorf("querying player: %w", err)
	}
	return &p, nil
}

// GetAllTeams returns a list of all NFL teams that have players
func (s *PlayerSyncer) GetAllTeams(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT DISTINCT team FROM nfl_players WHERE team IS NOT NULL AND team != '' ORDER BY team`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying teams: %w", err)
	}
	defer rows.Close()

	var teams []string
	for rows.Next() {
		var team string
		if err := rows.Scan(&team); err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, rows.Err()
}

// ExportPlayersJSON exports all players to a JSON format for the game
func (s *PlayerSyncer) ExportPlayersJSON(ctx context.Context) ([]byte, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, position, team, status
		FROM nfl_players
		WHERE status = 'Active' OR status = 'active'
		ORDER BY team, position, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying players: %w", err)
	}
	defer rows.Close()

	type exportPlayer struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Position string `json:"position"`
		Team     string `json:"team"`
	}

	var players []exportPlayer
	for rows.Next() {
		var p DBPlayer
		if err := rows.Scan(&p.ID, &p.Name, &p.Position, &p.Team, &p.Status); err != nil {
			return nil, fmt.Errorf("scanning player: %w", err)
		}
		players = append(players, exportPlayer{
			ID:       p.ID,
			Name:     p.Name,
			Position: p.Position,
			Team:     p.Team,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return json.MarshalIndent(players, "", "  ")
}
