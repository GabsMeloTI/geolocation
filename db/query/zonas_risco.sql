-- name: CreateZonaRisco :one
INSERT INTO zonas_risco (name, cep, lat, lng, radius, type, status)
VALUES ($1, $2, $3, $4, $5, $6, true)
RETURNING *;

-- name: UpdateZonaRisco :one
UPDATE zonas_risco
SET name = $2, cep = $3, lat = $4, lng = $5, radius = $6, type= $7
WHERE id = $1 and status = true
RETURNING *;

-- name: DeleteZonaRisco :exec
UPDATE zonas_risco
SET status = false
WHERE id = $1;

-- name: GetZonaRiscoById :one
SELECT * FROM zonas_risco WHERE id = $1 AND status = true;

-- name: GetAllZonasRisco :many
SELECT * FROM zonas_risco WHERE status = true;
