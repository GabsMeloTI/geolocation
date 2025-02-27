-- name: CreateUser :one
INSERT INTO users
(name, email, password, google_id, profile_picture)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

