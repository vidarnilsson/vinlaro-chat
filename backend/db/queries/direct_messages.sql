-- name: FindExistingDM :one
SELECT c.id FROM channels c
JOIN channel_members cm1 ON cm1.channel_id = c.id AND cm1.user_id = $1
JOIN channel_members cm2 ON cm2.channel_id = c.id AND cm2.user_id = $2
WHERE c.kind = 'dm'
LIMIT 1;

-- name: AddChannelMember :exec
INSERT INTO channel_members (channel_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetChannelMembers :many
SELECT u.id, u.username FROM users u
JOIN channel_members cm ON cm.user_id = u.id
WHERE cm.channel_id = $1;

-- name: GetUserDMs :many
SELECT
    c.id         AS channel_id,
    u.id         AS other_user_id,
    u.username   AS other_username
FROM channels c
JOIN channel_members cm_self  ON cm_self.channel_id  = c.id AND cm_self.user_id  = $1
JOIN channel_members cm_other ON cm_other.channel_id = c.id AND cm_other.user_id != $1
JOIN users u ON u.id = cm_other.user_id
WHERE c.kind = 'dm'
ORDER BY c.created_at DESC;

-- name: IsChannelMember :one
SELECT EXISTS (
    SELECT 1 FROM channel_members
    WHERE channel_id = $1 AND user_id = $2
) AS is_member;
