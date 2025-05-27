-- name: CreateRouteEnterprise :one
INSERT INTO public.route_enterprise
(id, origin, destination, waypoints, response, status, created_at, created_who, tenant_id, access_id)
VALUES(nextval('route_enterprise_id_seq'::regclass), $1, $2, $3, $4, true, now(), $5, $6, $7)
    RETURNING *;

-- name: DeleteRouteEnterprise :exec
UPDATE public.route_enterprise
SET status=false
WHERE id=$1 AND
      tenant_id=$2 AND
      access_id=$3;