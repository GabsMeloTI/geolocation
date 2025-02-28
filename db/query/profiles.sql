-- name: GetProfileById :one
SELECT *
FROM profiles
WHERE id=$1;