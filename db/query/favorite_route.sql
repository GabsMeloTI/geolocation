-- name: CreateFavoriteRoute :one
INSERT INTO public.favorite_route
(id, id_user, origin, destination, waypoints, response, created_at, status)
VALUES(nextval('favorite_route_id_seq'::regclass), $1, $2, $3, $4, $5, now(), true)
    RETURNING *;

-- name: RemoveFavorite :exec
UPDATE public.favorite_route
SET status=false, updated_at=now()
WHERE id = $1 AND
      id_user = $2;

-- name: GetFavoriteByUserId :many
SELECT *
FROM public.favorite_route
WHERE id_user = $1 AND
      status=true;
