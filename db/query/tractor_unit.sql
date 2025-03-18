-- name: CreateTractorUnit :one
INSERT INTO public.tractor_unit
(id, license_plate, driver_id, user_id, chassis, brand, model, manufacture_year, engine_power, unit_type, can_couple, height, axles, state, renavan, capacity, width, length, color, status, created_at)
VALUES(nextval('tractor_unit_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, true, now())
    RETURNING *;

-- name: UpdateTractorUnit :one
UPDATE public.tractor_unit
SET license_plate=$1, driver_id=$2, chassis=$3, brand=$4, model=$5, manufacture_year=$6, engine_power=$7, unit_type=$8, height=$9, axles=$11, state=$12, renavan=$13, capacity=$14, width=$15, length=$16, color=$17, updated_at=now()
WHERE id=$18 and
    user_id=$10
    RETURNING *;

-- name: DeleteTractorUnit :exec
UPDATE public.tractor_unit
SET status=false, updated_at=now()
WHERE id=$1 AND
      user_id=$2;

-- name: GetTractorUnitById :one
SELECT *
FROM public.tractor_unit
WHERE id=$1 AND
    status=true;

-- name: GetTractorUnitByUserId :many
SELECT *
FROM public.tractor_unit
WHERE user_id=$1 AND
    status=true;

-- name: GetOneTractorUnitByUserId :one
SELECT *
FROM public.tractor_unit
WHERE user_id=$1 AND
    status=true;
