-- name: CreateSavedRoutes :one
INSERT INTO public.saved_routes
(id, origin, destination, waypoints, response, created_at)
VALUES(nextval('saved_routes_id_seq'::regclass), $1, $2, $3, $4, now())
    RETURNING *;


-- name: GetSavedRoutes :one
SELECT *
FROM public.saved_routes
WHERE origin = $1 AND
      destination = $2 AND
      waypoints = $3;

-- name: AddSavedRoutesFavorite :exec
UPDATE public.saved_routes
SET favorite=true, upated_at=now()
WHERE id = $1;

-- name: GetSavedRouteById :one
SELECT *
FROM public.saved_routes
WHERE ID = $1;