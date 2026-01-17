-- name: CreateWeeklyDraft :one
INSERT INTO weekly_drafts (season_id, member_id, week, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWeeklyDraft :one
SELECT * FROM weekly_drafts WHERE season_id = $1 AND member_id = $2 AND week = $3;

-- name: GetWeeklyDraftByID :one
SELECT * FROM weekly_drafts WHERE id = $1;

-- name: GetWeeklyDraftsBySeasonWeek :many
SELECT wd.*, lm.team_name, u.display_name
FROM weekly_drafts wd
INNER JOIN league_members lm ON wd.member_id = lm.id
INNER JOIN users u ON lm.user_id = u.id
WHERE wd.season_id = $1 AND wd.week = $2
ORDER BY wd.started_at;

-- name: StartWeeklyDraft :one
UPDATE weekly_drafts
SET status = 'in_progress', started_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CompleteWeeklyDraft :one
UPDATE weekly_drafts
SET status = 'complete', completed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateDraftPick :one
INSERT INTO draft_picks (weekly_draft_id, pick_number, nfl_player_id, player_name, position, team_drawn)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDraftPicks :many
SELECT * FROM draft_picks WHERE weekly_draft_id = $1 ORDER BY pick_number;

-- name: GetWeeklyResult :one
SELECT * FROM weekly_results WHERE season_id = $1 AND member_id = $2 AND week = $3;

-- name: GetWeeklyResultsBySeasonWeek :many
SELECT wr.*, lm.team_name, u.display_name
FROM weekly_results wr
INNER JOIN league_members lm ON wr.member_id = lm.id
INNER JOIN users u ON lm.user_id = u.id
WHERE wr.season_id = $1 AND wr.week = $2
ORDER BY wr.total_points DESC;

-- name: CreateWeeklyResult :one
INSERT INTO weekly_results (season_id, member_id, week, total_points, weekly_rank, is_weekly_winner, calculated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING *;

-- name: UpdateWeeklyResult :one
UPDATE weekly_results
SET total_points = $4,
    weekly_rank = $5,
    is_weekly_winner = $6,
    calculated_at = NOW()
WHERE season_id = $1 AND member_id = $2 AND week = $3
RETURNING *;

-- name: UpsertWeeklyResult :one
INSERT INTO weekly_results (season_id, member_id, week, total_points, weekly_rank, is_weekly_winner, calculated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (season_id, member_id, week) DO UPDATE SET
    total_points = EXCLUDED.total_points,
    weekly_rank = EXCLUDED.weekly_rank,
    is_weekly_winner = EXCLUDED.is_weekly_winner,
    calculated_at = NOW()
RETURNING *;
