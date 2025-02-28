-- name: Login :one
SELECT
    *
FROM
    users
WHERE
    email = $1 and
    google_id = $2 and
    status = true;
