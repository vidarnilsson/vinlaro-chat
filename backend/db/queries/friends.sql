-- name: SendFriendRequest :one
INSERT INTO friendships (requester_id, addressee_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetFriendship :one
SELECT * FROM friendships
WHERE (requester_id = $1 AND addressee_id = $2)
   OR (requester_id = $2 AND addressee_id = $1)
LIMIT 1;

-- name: GetFriendshipByID :one
SELECT * FROM friendships WHERE id = $1;

-- name: UpdateFriendshipStatus :one
UPDATE friendships SET status = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: DeleteFriendship :exec
DELETE FROM friendships WHERE id = $1;

-- name: GetPendingFriendRequests :many
SELECT f.id, f.requester_id, u.username AS requester_username, f.created_at
FROM friendships f
JOIN users u ON u.id = f.requester_id
WHERE f.addressee_id = $1 AND f.status = 'pending'
ORDER BY f.created_at DESC;

-- name: GetFriends :many
SELECT
    f.id AS friendship_id,
    CASE
        WHEN f.requester_id = $1 THEN f.addressee_id
        ELSE f.requester_id
    END AS friend_id,
    u.username AS friend_username,
    f.status
FROM friendships f
JOIN users u ON u.id = CASE
    WHEN f.requester_id = $1 THEN f.addressee_id
    ELSE f.requester_id
END
WHERE (f.requester_id = $1 OR f.addressee_id = $1)
  AND f.status = 'accepted'
ORDER BY u.username ASC;
