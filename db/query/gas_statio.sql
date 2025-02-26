-- name: CreateGasStations :one
INSERT INTO public.gas_station
(id, name, latitude, longitude, address_name, municipio, specific_point)
VALUES(nextval('gas_station_id_seq'::regclass), $1, $2, $3, $4, $5, $6)
    RETURNING *;


-- name: GetGasStation :many
SELECT id, latitude, longitude, address_name, municipio, specific_point, name
FROM gas_station
WHERE
    CAST(latitude AS FLOAT) BETWEEN CAST($1 AS FLOAT) - $3 AND CAST($1 AS FLOAT) + $3
  AND CAST(longitude AS FLOAT) BETWEEN CAST($2 AS FLOAT) - $3 AND CAST($2 AS FLOAT) + $3;

-- name: GetGasStationsByBoundingBox :many
SELECT id, latitude, longitude, address_name, municipio, specific_point, name
FROM gas_station
WHERE CAST(latitude AS FLOAT) BETWEEN CAST($1 AS FLOAT) AND CAST($2 AS FLOAT)
  AND CAST(longitude AS FLOAT) BETWEEN CAST($3 AS FLOAT) AND CAST($4 AS FLOAT);
