-- name: CreateRouteHist :one
INSERT INTO public.route_hist
(id, id_user, origin, destination, waypoints, response, is_public, number_request, created_at)
VALUES(nextval('route_hist_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7,now())
    RETURNING *;

-- name: UpdateRouteHistCount :one
UPDATE route_hist
SET number_request = number_request + 1
WHERE id = $1
    RETURNING number_request;



-- name: GetRouteHistByUnique :one
SELECT *
FROM route_hist
WHERE id_user = $1
  AND origin = $2
  AND destination = $3
  AND waypoints = $4
  AND is_public = $5
    LIMIT 1;
