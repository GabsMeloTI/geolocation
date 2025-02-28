-- name: CreateTractorUnit :one
INSERT INTO public.tractor_unit
(id, license_plate, driver_id, user_id, chassis, brand, model, manufacture_year, engine_power, unit_type, can_couple, height, axles, status, created_at)
VALUES(nextval('tractor_unit_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, true, now())
    RETURNING *;

-- name: UpdateTractorUnit :one
UPDATE public.tractor_unit
SET license_plate=$1, driver_id=$2, chassis=$3, brand=$4, model=$5, manufacture_year=$6, engine_power=$7, unit_type=$8, height=$9, user_id=$10, axles=$11, updated_at=now()
WHERE id=$12
    RETURNING *;

-- name: DeleteTractorUnit :exec
UPDATE public.tractor_unit
SET status=false, updated_at=now()
WHERE id=$1;

-- name: GetTractorUnitById :one
SELECT *
FROM public.tractor_unit
WHERE id=$1;


