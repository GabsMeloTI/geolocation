-- name: CreateLocation :one
INSERT INTO public.locations
(id, "type", address, id_provider_info, color, created_at, access_id, tenant_id)
VALUES(nextval('locations_id_seq'::regclass), $1, $2, $3, $4, now(), $5, $6)
    RETURNING *;

-- name: UpdateLocation :one
UPDATE public.locations
SET "type"=$1, address=$2, id_provider_info=$6, color=$7, updated_at=now()
WHERE id=$3 AND
    access_id=$4 AND
    tenant_id=$5
    RETURNING *;

-- name: DeleteLocation :exec
DELETE FROM public.locations
WHERE id=$1 AND
    access_id=$2 AND
    tenant_id=$3;


-- name: CreateArea :one
INSERT INTO public.areas
(id, locations_id, latitude, longitude, description, created_at)
VALUES(nextval('areas_id_seq'::regclass), $1, $2, $3, $4, now())
    RETURNING *;

-- name: UpdateArea :one
UPDATE public.areas
SET locations_id=$1, latitude=$2, longitude=$3, description=$4, created_at=now()
WHERE id=$5
    RETURNING *;

-- name: DeleteArea :exec
DELETE FROM public.areas
WHERE locations_id=$1;

-- name: GetLocationByOrg :many
SELECT l.id, l.type, l.address, l.id_provider_info, l.color, l.created_at, l.updated_at, l.access_id, l.tenant_id
FROM public.locations l
WHERE id_provider_info=$3 AND
      access_id=$1 AND
      tenant_id=$2;

-- name: GetAreasByOrg :many
SELECT a.id, a.locations_id, a.latitude, a.longitude, a.description
FROM public.areas a
WHERE locations_id=$1;


-- name: GetLocationByName :many
SELECT *
FROM public.locations l
WHERE id_provider_info=$3 AND
      type=$4 AND
      access_id=$1 AND
      tenant_id=$2;


-- name: GetLocationByNameExcludingID :one
SELECT *
FROM public.locations
WHERE access_id  = $1
  AND tenant_id  = $2
  AND type      = $3
  AND id         <> $4
  AND id_provider_info = $5;
