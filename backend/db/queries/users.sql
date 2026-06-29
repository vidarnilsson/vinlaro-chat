-- queries/users.sql

-- name: CreateUser :one
INSERT INTO users (username, email, password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: SearchUsers :many
SELECT id, username FROM users
WHERE username ILIKE $1
AND id != $2
LIMIT 10;
