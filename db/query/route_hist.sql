-- name: CreateRouteHist :one
INSERT INTO public.route_hist
(id, id_token_hist, origin, destination, waypoints, response, created_at)
VALUES(nextval('route_hist_id_seq'::regclass), $1, $2, $3, $4, $5, now())
    RETURNING *;