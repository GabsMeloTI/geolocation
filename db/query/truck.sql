-- name: CreateTruck :one
INSERT INTO public.truck
(id, tractor_unit_id, trailer_id, driver_id)
VALUES(nextval('truck_id_seq'::regclass), $1, $2, $3)
    RETURNING *;