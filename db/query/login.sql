-- name: Login :one
SELECT
    *
FROM
    users
WHERE
    email = $1 and
    google_id = $2 and
    status = true;


-- name: NewCreateUser :one
INSERT INTO public.users
("name", email, "password", created_at, profile_id, "document", phone, google_id, profile_picture, status)
VALUES( $1, $2, $3, CURRENT_TIMESTAMP, $4, $5, $6, $7, $8, true)
    returning *;
