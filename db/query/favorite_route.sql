-- name: CreateFavoriteRoute :one
INSERT INTO public.favorite_route
(id, tolls_id, response, user_organization, created_who, created_at)
VALUES(nextval('favorite_route_id_seq'::regclass), $1, $2, $3, $4, now())
    RETURNING *;

