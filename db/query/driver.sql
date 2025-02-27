-- name: CreateDriver :one
INSERT INTO public.driver
(id, user_id, birth_date, cpf, license_number, license_category, license_expiration_date, state, city, neighborhood, street, street_number, phone, status, created_at)
VALUES(nextval('driver_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
       $12, true, now())
    RETURNING *;

-- name: UpdateDriver :one
UPDATE public.driver
SET user_id=$2, birth_date=$3, license_category=$4, license_expiration_date=$5, state=$6, city=$7, neighborhood=$8, street=$9, street_number=$10, phone=$11, updated_at=now()
WHERE id=$1
    RETURNING *;

-- name: DeleteDriver :exec
UPDATE public.driver
SET status=false, updated_at=now()
WHERE id=$1;

-- name: GetDriverById :one
SELECT *
FROM public.driver
WHERE id=$1;
