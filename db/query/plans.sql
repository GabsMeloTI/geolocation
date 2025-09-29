-- name: GetPlansById :one
SELECT *
FROM public."plans"
WHERE id=$1;


