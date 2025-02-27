-- name: CreateFavoriteRoute :one
INSERT INTO public.favorite_route
(id, id_user, origin, destination, waypoints, response, created_who, created_at)
VALUES(nextval('favorite_route_id_seq'::regclass), $1, $2, $3, $4, $5, $6, now())
    RETURNING *;

-- name: RemoveFavorite :exec
DELETE
FROM public.favorite_route
WHERE id = $1 AND
    id_user = $2;