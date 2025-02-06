-- name: CreateSavedRoutes :one
INSERT INTO public.saved_routes
(id, origin, destination, waypoints, request, response, created_at, expired_at)
VALUES(nextval('saved_routes_id_seq'::regclass), $1, $2, $3, $4, $5,now(), $6)
    RETURNING *;

-- name: GetSavedRoutes :one
SELECT *
FROM public.saved_routes
WHERE origin = $1 AND
      destination = $2 AND
      waypoints = $3 ;

-- name: GetSavedRouteById :one
SELECT *
FROM public.saved_routes
WHERE ID = $1;