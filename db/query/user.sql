-- name: CreateUser :one
INSERT INTO users
(name, email, password, google_id, profile_picture, status, phone, document, profile_id, driver_id)
VALUES ($1, $2, $3, $4, $5, true, $6, $7, $8, $9)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 and status;

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


-- name: UpdateUserPassword :exec
UPDATE users
SET password = $1
WHERE id = $2;


-- name: UpdateUserPersonalInfo :one
UPDATE users
SET name=$1, "document" = $2, email = $3, phone = $4, updated_at = now(), secondary_contact = $6, date_of_birth = $7
WHERE id = $5
    RETURNING *;

-- name: UpdateUserAddress :one
UPDATE users
SET complement=$1, state = $2, city = $3, neighborhood = $4, street = $5, street_number=$6, cep = $7, updated_at = now()
WHERE id = $8
    RETURNING *;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: CreateHistoryRecoverPassword :exec
INSERT INTO history_recover_password 
(user_id, email, token)
VALUES
($1, $2, $3);

-- name: UpdatePasswordByUserId :exec
UPDATE users
SET password = $1
WHERE id = $2;

-- name: UpdateHistoryRecoverPassword :exec
update history_recover_password
set date_update_password = now()
where token = $1;


-- name: UpdateProfilePictureByUserId :exec
UPDATE users
SET profile_picture = $1
WHERE id = $2;