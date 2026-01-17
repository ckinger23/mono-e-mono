package nfl

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ScoringService handles calculating and storing weekly scores
type ScoringService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewScoringService creates a new scoring service
func NewScoringService(pool *pgxpool.Pool, logger *slog.Logger) *ScoringService {
	return &ScoringService{
		pool:   pool,
		logger: logger,
	}
}

// ScoringResult contains the results of calculating scores
type ScoringResult struct {
	SeasonID      string            `json:"season_id"`
	Week          int               `json:"week"`
	MembersScored int               `json:"members_scored"`
	WeeklyWinner  *MemberScore      `json:"weekly_winner,omitempty"`
	AllScores     []MemberScore     `json:"all_scores"`
	Errors        int               `json:"errors"`
}

// MemberScore represents a member's weekly score
type MemberScore struct {
	MemberID    string  `json:"member_id"`
	TeamName    string  `json:"team_name"`
	UserID      string  `json:"user_id"`
	TotalPoints float64 `json:"total_points"`
	WeeklyRank  int     `json:"weekly_rank"`
}

// CalculateWeeklyScores calculates and stores scores for all members in a season for a given week
func (s *ScoringService) CalculateWeeklyScores(ctx context.Context, seasonID string, week int) (*ScoringResult, error) {
	s.logger.Info("calculating weekly scores", "season_id", seasonID, "week", week)

	// Get season info to determine scoring type and year
	var scoringType string
	var year int
	err := s.pool.QueryRow(ctx,
		`SELECT l.scoring_type, s.year
		FROM seasons s
		INNER JOIN leagues l ON s.league_id = l.id
		WHERE s.id = $1`,
		seasonID,
	).Scan(&scoringType, &year)
	if err != nil {
		return nil, fmt.Errorf("fetching season info: %w", err)
	}

	// Get all completed drafts for this week
	rows, err := s.pool.Query(ctx,
		`SELECT wd.id, wd.member_id, lm.team_name, lm.user_id
		FROM weekly_drafts wd
		INNER JOIN league_members lm ON wd.member_id = lm.id
		WHERE wd.season_id = $1 AND wd.week = $2 AND wd.status = 'complete'`,
		seasonID, week,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching drafts: %w", err)
	}
	defer rows.Close()

	type draftInfo struct {
		draftID  string
		memberID string
		teamName string
		userID   string
	}
	var drafts []draftInfo

	for rows.Next() {
		var d draftInfo
		if err := rows.Scan(&d.draftID, &d.memberID, &d.teamName, &d.userID); err != nil {
			return nil, fmt.Errorf("scanning draft: %w", err)
		}
		drafts = append(drafts, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating drafts: %w", err)
	}

	result := &ScoringResult{
		SeasonID: seasonID,
		Week:     week,
	}

	// Calculate score for each draft
	for _, draft := range drafts {
		score, err := s.calculateDraftScore(ctx, draft.draftID, year, week, scoringType)
		if err != nil {
			s.logger.Error("failed to calculate draft score",
				"draft_id", draft.draftID,
				"error", err,
			)
			result.Errors++
			continue
		}

		result.AllScores = append(result.AllScores, MemberScore{
			MemberID:    draft.memberID,
			TeamName:    draft.teamName,
			UserID:      draft.userID,
			TotalPoints: score,
		})
		result.MembersScored++
	}

	// Sort by points descending and assign ranks
	sort.Slice(result.AllScores, func(i, j int) bool {
		return result.AllScores[i].TotalPoints > result.AllScores[j].TotalPoints
	})

	for i := range result.AllScores {
		result.AllScores[i].WeeklyRank = i + 1
	}

	// Store results and update standings
	if err := s.storeResults(ctx, result); err != nil {
		return nil, fmt.Errorf("storing results: %w", err)
	}

	if len(result.AllScores) > 0 {
		result.WeeklyWinner = &result.AllScores[0]
	}

	s.logger.Info("weekly scores calculated",
		"season_id", seasonID,
		"week", week,
		"members_scored", result.MembersScored,
	)

	return result, nil
}

// calculateDraftScore sums up points for all picks in a draft
func (s *ScoringService) calculateDraftScore(ctx context.Context, draftID string, year, week int, scoringType string) (float64, error) {
	// Get all picks for this draft
	rows, err := s.pool.Query(ctx,
		`SELECT nfl_player_id FROM draft_picks WHERE weekly_draft_id = $1`,
		draftID,
	)
	if err != nil {
		return 0, fmt.Errorf("fetching picks: %w", err)
	}
	defer rows.Close()

	var totalPoints float64
	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return 0, fmt.Errorf("scanning pick: %w", err)
		}

		// Get player's points for this week
		var points float64
		var column string
		switch scoringType {
		case "standard":
			column = "points_standard"
		case "half_ppr":
			column = "points_half_ppr"
		default:
			column = "points_ppr"
		}

		query := fmt.Sprintf(
			`SELECT COALESCE(%s, 0) FROM nfl_weekly_stats WHERE player_id = $1 AND year = $2 AND week = $3`,
			column,
		)
		err := s.pool.QueryRow(ctx, query, playerID, year, week).Scan(&points)
		if err != nil {
			// Player might not have stats (bye week, injury, etc.)
			s.logger.Debug("no stats for player", "player_id", playerID, "week", week)
			continue
		}

		totalPoints += points
	}

	return totalPoints, rows.Err()
}

// storeResults saves weekly results and updates standings
func (s *ScoringService) storeResults(ctx context.Context, result *ScoringResult) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert/update weekly results
	for _, score := range result.AllScores {
		isWinner := score.WeeklyRank == 1

		_, err := tx.Exec(ctx,
			`INSERT INTO weekly_results (season_id, member_id, week, total_points, weekly_rank, is_weekly_winner, calculated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (season_id, member_id, week) DO UPDATE SET
				total_points = EXCLUDED.total_points,
				weekly_rank = EXCLUDED.weekly_rank,
				is_weekly_winner = EXCLUDED.is_weekly_winner,
				calculated_at = NOW()`,
			result.SeasonID, score.MemberID, result.Week, score.TotalPoints, score.WeeklyRank, isWinner,
		)
		if err != nil {
			return fmt.Errorf("inserting weekly result: %w", err)
		}
	}

	// Update standings (aggregate all weekly results)
	_, err = tx.Exec(ctx,
		`INSERT INTO standings (season_id, member_id, weekly_wins, total_points, best_week, weeks_played)
		SELECT
			season_id,
			member_id,
			COUNT(*) FILTER (WHERE is_weekly_winner) as weekly_wins,
			SUM(total_points) as total_points,
			MAX(total_points) as best_week,
			COUNT(*) as weeks_played
		FROM weekly_results
		WHERE season_id = $1
		GROUP BY season_id, member_id
		ON CONFLICT (season_id, member_id) DO UPDATE SET
			weekly_wins = EXCLUDED.weekly_wins,
			total_points = EXCLUDED.total_points,
			best_week = EXCLUDED.best_week,
			weeks_played = EXCLUDED.weeks_played`,
		result.SeasonID,
	)
	if err != nil {
		return fmt.Errorf("updating standings: %w", err)
	}

	// Update current_rank in standings
	_, err = tx.Exec(ctx,
		`UPDATE standings s
		SET current_rank = ranked.rank
		FROM (
			SELECT member_id, RANK() OVER (ORDER BY total_points DESC) as rank
			FROM standings
			WHERE season_id = $1
		) ranked
		WHERE s.season_id = $1 AND s.member_id = ranked.member_id`,
		result.SeasonID,
	)
	if err != nil {
		return fmt.Errorf("updating ranks: %w", err)
	}

	return tx.Commit(ctx)
}

// GetWeeklyResults returns stored results for a week
func (s *ScoringService) GetWeeklyResults(ctx context.Context, seasonID string, week int) ([]MemberScore, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT wr.member_id, lm.team_name, lm.user_id, wr.total_points, wr.weekly_rank
		FROM weekly_results wr
		INNER JOIN league_members lm ON wr.member_id = lm.id
		WHERE wr.season_id = $1 AND wr.week = $2
		ORDER BY wr.weekly_rank`,
		seasonID, week,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching results: %w", err)
	}
	defer rows.Close()

	var scores []MemberScore
	for rows.Next() {
		var s MemberScore
		if err := rows.Scan(&s.MemberID, &s.TeamName, &s.UserID, &s.TotalPoints, &s.WeeklyRank); err != nil {
			return nil, fmt.Errorf("scanning result: %w", err)
		}
		scores = append(scores, s)
	}

	return scores, rows.Err()
}

// StandingsEntry represents a member's season standings
type StandingsEntry struct {
	MemberID    string  `json:"member_id"`
	TeamName    string  `json:"team_name"`
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	WeeklyWins  int     `json:"weekly_wins"`
	TotalPoints float64 `json:"total_points"`
	BestWeek    float64 `json:"best_week"`
	WeeksPlayed int     `json:"weeks_played"`
	CurrentRank int     `json:"current_rank"`
}

// GetStandings returns current standings for a season
func (s *ScoringService) GetStandings(ctx context.Context, seasonID string) ([]StandingsEntry, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT s.member_id, lm.team_name, lm.user_id, u.display_name,
			s.weekly_wins, s.total_points, s.best_week, s.weeks_played, s.current_rank
		FROM standings s
		INNER JOIN league_members lm ON s.member_id = lm.id
		INNER JOIN users u ON lm.user_id = u.id
		WHERE s.season_id = $1
		ORDER BY s.current_rank`,
		seasonID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching standings: %w", err)
	}
	defer rows.Close()

	var entries []StandingsEntry
	for rows.Next() {
		var e StandingsEntry
		if err := rows.Scan(&e.MemberID, &e.TeamName, &e.UserID, &e.DisplayName,
			&e.WeeklyWins, &e.TotalPoints, &e.BestWeek, &e.WeeksPlayed, &e.CurrentRank); err != nil {
			return nil, fmt.Errorf("scanning standing: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// RecalculateAllWeeks recalculates scores for all completed weeks in a season
func (s *ScoringService) RecalculateAllWeeks(ctx context.Context, seasonID string) error {
	// Get current week from season
	var currentWeek int
	err := s.pool.QueryRow(ctx,
		`SELECT current_week FROM seasons WHERE id = $1`,
		seasonID,
	).Scan(&currentWeek)
	if err != nil {
		return fmt.Errorf("fetching current week: %w", err)
	}

	// Recalculate each week
	for week := 1; week <= currentWeek; week++ {
		if _, err := s.CalculateWeeklyScores(ctx, seasonID, week); err != nil {
			s.logger.Error("failed to recalculate week", "week", week, "error", err)
		}
	}

	return nil
}
