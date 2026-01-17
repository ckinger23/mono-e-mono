-- name: UpsertNFLPlayer :one
INSERT INTO nfl_players (id, name, position, team, status, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    position = EXCLUDED.position,
    team = EXCLUDED.team,
    status = EXCLUDED.status,
    updated_at = NOW()
RETURNING *;

-- name: GetNFLPlayer :one
SELECT * FROM nfl_players WHERE id = $1;

-- name: GetNFLPlayersByTeam :many
SELECT * FROM nfl_players WHERE team = $1 ORDER BY position, name;

-- name: GetNFLPlayersByPosition :many
SELECT * FROM nfl_players WHERE position = $1 ORDER BY team, name;

-- name: GetAllNFLPlayers :many
SELECT * FROM nfl_players ORDER BY team, position, name;

-- name: UpsertNFLWeeklyStats :one
INSERT INTO nfl_weekly_stats (player_id, year, week, points_ppr, points_standard, points_half_ppr, stats)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (player_id, year, week) DO UPDATE SET
    points_ppr = EXCLUDED.points_ppr,
    points_standard = EXCLUDED.points_standard,
    points_half_ppr = EXCLUDED.points_half_ppr,
    stats = EXCLUDED.stats
RETURNING *;

-- name: GetNFLWeeklyStats :one
SELECT * FROM nfl_weekly_stats WHERE player_id = $1 AND year = $2 AND week = $3;

-- name: GetNFLWeeklyStatsByWeek :many
SELECT nws.*, np.name, np.position, np.team
FROM nfl_weekly_stats nws
INNER JOIN nfl_players np ON nws.player_id = np.id
WHERE nws.year = $1 AND nws.week = $2
ORDER BY nws.points_ppr DESC;
