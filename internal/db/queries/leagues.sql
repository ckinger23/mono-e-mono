-- name: CreateLeague :one
INSERT INTO leagues (name, invite_code, commissioner_id, max_members, scoring_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLeagueByID :one
SELECT * FROM leagues WHERE id = $1;

-- name: GetLeagueByInviteCode :one
SELECT * FROM leagues WHERE invite_code = $1;

-- name: GetLeaguesByUserID :many
SELECT l.* FROM leagues l
INNER JOIN league_members lm ON l.id = lm.league_id
WHERE lm.user_id = $1
ORDER BY l.created_at DESC;

-- name: UpdateLeague :one
UPDATE leagues
SET name = COALESCE($2, name),
    max_members = COALESCE($3, max_members),
    scoring_type = COALESCE($4, scoring_type)
WHERE id = $1 AND commissioner_id = $5
RETURNING *;

-- name: DeleteLeague :exec
DELETE FROM leagues WHERE id = $1 AND commissioner_id = $2;

-- name: CreateLeagueMember :one
INSERT INTO league_members (league_id, user_id, team_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLeagueMember :one
SELECT * FROM league_members WHERE league_id = $1 AND user_id = $2;

-- name: GetLeagueMemberByID :one
SELECT * FROM league_members WHERE id = $1;

-- name: GetLeagueMembers :many
SELECT lm.*, u.display_name, u.avatar_url
FROM league_members lm
INNER JOIN users u ON lm.user_id = u.id
WHERE lm.league_id = $1
ORDER BY lm.joined_at;

-- name: GetLeagueMemberCount :one
SELECT COUNT(*) FROM league_members WHERE league_id = $1;

-- name: UpdateLeagueMemberTeamName :one
UPDATE league_members
SET team_name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteLeagueMember :exec
DELETE FROM league_members WHERE id = $1;
