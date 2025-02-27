-- name: CreateUser :one
INSERT INTO users
(name, email, password, google_id, profile_picture, status)
VALUES ($1, $2, $3, $4, $5, true)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: DeleteUserById :exec
UPDATE users SET status = false WHERE id = $1;

-- name: UpdateUserById :one
UPDATE users
SET name = $1,
    profile_picture = $2,
    state = $3,
    city = $4,
    neighborhood = $5,
    street = $6,
    street_number = $7,
    phone = $8,
    updated_at = now()

WHERE id = $9
    RETURNING *;

