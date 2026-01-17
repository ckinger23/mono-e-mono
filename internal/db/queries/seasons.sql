-- name: CreateSeason :one
INSERT INTO seasons (league_id, year, current_week, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSeasonByID :one
SELECT * FROM seasons WHERE id = $1;

-- name: GetActiveSeasonByLeague :one
SELECT * FROM seasons WHERE league_id = $1 AND status = 'active';

-- name: GetSeasonsByLeague :many
SELECT * FROM seasons WHERE league_id = $1 ORDER BY year DESC;

-- name: UpdateSeasonWeek :one
UPDATE seasons
SET current_week = $2
WHERE id = $1
RETURNING *;

-- name: CompleteSeason :one
UPDATE seasons
SET status = 'complete', champion_member_id = $2
WHERE id = $1
RETURNING *;

-- name: GetStandingsBySeason :many
SELECT s.*, lm.team_name, u.display_name
FROM standings s
INNER JOIN league_members lm ON s.member_id = lm.id
INNER JOIN users u ON lm.user_id = u.id
WHERE s.season_id = $1
ORDER BY s.total_points DESC;

-- name: GetStandingByMember :one
SELECT * FROM standings WHERE season_id = $1 AND member_id = $2;

-- name: CreateStanding :one
INSERT INTO standings (season_id, member_id, weekly_wins, total_points, best_week, weeks_played, current_rank)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateStanding :one
UPDATE standings
SET weekly_wins = $3,
    total_points = $4,
    best_week = $5,
    weeks_played = $6,
    current_rank = $7
WHERE season_id = $1 AND member_id = $2
RETURNING *;

-- name: UpsertStanding :one
INSERT INTO standings (season_id, member_id, weekly_wins, total_points, best_week, weeks_played, current_rank)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (season_id, member_id) DO UPDATE SET
    weekly_wins = EXCLUDED.weekly_wins,
    total_points = EXCLUDED.total_points,
    best_week = EXCLUDED.best_week,
    weeks_played = EXCLUDED.weeks_played,
    current_rank = EXCLUDED.current_rank
RETURNING *;
