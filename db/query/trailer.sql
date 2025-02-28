-- name: CreateTrailer :one
INSERT INTO public.trailer
(id, license_plate, user_id, chassis, body_type, load_capacity, length, width, height, axles, status, created_at)
VALUES(nextval('trailer_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9,  true, now())
    RETURNING *;


-- name: UpdateTrailer :one
UPDATE public.trailer
SET license_plate=$1, chassis=$2, body_type=$3, load_capacity=$4, length=$5, width=$6, height=$7, axles=$8, user_id=$9, updated_at=now()
WHERE id=$10
    RETURNING *;


-- name: DeleteTrailer :exec
UPDATE public.trailer
SET status=false, updated_at=now()
WHERE id=$1;

-- name: GetTrailerById :one
SELECT *
FROM public.trailer
WHERE id=$1;