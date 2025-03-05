-- name: CreatePlans :one
INSERT INTO public."plans"
(id, "name", price, duration, annual)
VALUES(nextval('plans_id_seq'::regclass), $1, $2, $3, $4)
    RETURNING *;

-- name: UpdatePlans :one
UPDATE public."plans"
SET "name"=$2, price=$3, duration=$4, annual=$5
WHERE id=$1
    RETURNING *;


