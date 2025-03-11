-- name: CreateActiveFreight :exec
INSERT INTO active_freights
(advertisement_id, advertisement_user_id, latitude, longitude, duration, distance, driver_name, tractor_unit_license_plate, trailer_license_plate, updated_at)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, now());


-- name: UpdateActiveFreight :exec
UPDATE active_freights
SET latitude = $1,
    longitude = $2,
    duration = $3,
    distance = $4,
    updated_at = now()
WHERE id = $5;

-- name: GetAllActiveFreights :many
SELECT * FROM active_freights
WHERE advertisement_user_id = $1;

-- name: GetActiveFreight :one
SELECT * FROM active_freights
WHERE advertisement_id = $1;