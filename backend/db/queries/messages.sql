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
