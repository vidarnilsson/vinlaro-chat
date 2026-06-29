-- queries/channels.sql

-- name: CreateChannel :one
INSERT INTO channels (name, description, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListChannels :many
SELECT * FROM channels
ORDER BY name ASC;

-- name: GetChannelByID :one
SELECT * FROM channels
WHERE id = $1;

-- name: CreatePrivateChannel :one
INSERT INTO channels (name, description, created_by, kind)
VALUES ($1, $2, $3, 'private')
RETURNING *;

-- name: GetUserChannels :many
SELECT c.* FROM channels c
WHERE c.kind = 'public'
UNION
SELECT c.* FROM channels c
JOIN channel_members cm ON cm.channel_id = c.id
WHERE c.kind = 'private' AND cm.user_id = $1
ORDER BY name ASC;

-- name: AddChannelMemberWithRole :exec
INSERT INTO channel_members (channel_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (channel_id, user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: GetChannelMemberRole :one
SELECT role FROM channel_members
WHERE channel_id = $1 AND user_id = $2;


-- queries/messages.sql

-- name: CreateMessage :one
INSERT INTO messages (channel_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateMessageWithID :one
INSERT INTO messages (id, channel_id, user_id, content, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMessagesByChannel :many
SELECT
    m.id,
    m.content,
    m.created_at,
    m.channel_id,
    u.id       AS user_id,
    u.username AS username
FROM messages m
JOIN users u ON u.id = m.user_id
WHERE m.channel_id = $1
ORDER BY m.created_at ASC
LIMIT $2;
