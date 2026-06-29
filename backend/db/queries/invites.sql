-- name: CreateChannelInvite :one
INSERT INTO channel_invites (channel_id, inviter_id, invitee_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetChannelInvite :one
SELECT * FROM channel_invites WHERE id = $1;

-- name: GetPendingInvites :many
SELECT
    ci.id,
    ci.channel_id,
    c.name      AS channel_name,
    u.username  AS inviter_username,
    ci.created_at
FROM channel_invites ci
JOIN channels c ON c.id = ci.channel_id
JOIN users u    ON u.id = ci.inviter_id
WHERE ci.invitee_id = $1 AND ci.status = 'pending'
ORDER BY ci.created_at DESC;

-- name: UpdateInviteStatus :one
UPDATE channel_invites SET status = $1
WHERE id = $2
RETURNING *;

-- name: GetExistingInvite :one
SELECT id FROM channel_invites
WHERE channel_id = $1 AND invitee_id = $2 AND status = 'pending'
LIMIT 1;
