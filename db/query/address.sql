-- name: FindAddressesByQuery :many
SELECT
    s.id AS street_id,
    s.name AS street_name,
    n.name AS neighborhood_name,
    n.lat AS neighborhood_lat,
    n.lon AS neighborhood_lon,
    c.name AS city_name,
    c.lat AS city_lat,
    c.lon AS city_lon,
    st.uf AS state_uf,
    st.lat AS state_lat,
    st.lon AS state_lon,
    a.id AS address_id,
    a.number,
    a.cep,
    a.lat,
    a.lon
FROM streets s
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
         LEFT JOIN addresses a ON a.street_id = s.id
WHERE
    (
        COALESCE(
                NULLIF(sqlc.arg('street'), ''),
                NULLIF(sqlc.arg('city'), ''),
                NULLIF(sqlc.arg('state'), ''),
                NULLIF(sqlc.arg('neighborhood'), '')
        ) IS NOT NULL
        )

  AND (
    sqlc.arg('street') = '' OR
    regexp_replace(lower(unaccent(s.name)), '[z]', 's', 'gi')
        LIKE '%' || regexp_replace(lower(unaccent(sqlc.arg('street'))), '[z]', 's', 'gi') || '%'
    )

  AND (
    sqlc.arg('city') = '' OR
    c.search_vector @@ plainto_tsquery('portuguese', sqlc.arg('city'))
    )

  AND (
    sqlc.arg('state') = '' OR
    st.search_vector @@ plainto_tsquery('portuguese', sqlc.arg('state'))
    )

  AND (
    sqlc.arg('neighborhood') = '' OR
    n.search_vector @@ plainto_tsquery('portuguese', sqlc.arg('neighborhood'))
    )

  AND (
    sqlc.arg('number') = '' OR
    a.number ILIKE sqlc.arg('number') || '%'
    )

ORDER BY s.id
LIMIT 5;

-- name: FindAddressesByLatLon :many
SELECT
    s.id AS street_id,
    s.name AS street_name,
    n.name AS neighborhood_name,
    n.lat AS neighborhood_lat,
    n.lon AS neighborhood_lon,
    c.name AS city_name,
    c.lat AS city_lat,
    c.lon AS city_lon,
    st.uf AS state_uf,
    st.lat AS state_lat,
    st.lon AS state_lon,
    a.id AS address_id,
    a.number,
    a.cep,
    a.lat,
    a.lon
FROM addresses a
         JOIN streets s ON a.street_id = s.id
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
ORDER BY (a.lat - $1) * (a.lat - $1) + (a.lon - $2) * (a.lon - $2) ASC
limit 5;

-- name: FindAddressesByCEP :many
SELECT
    s.id AS street_id,
    s.name AS street_name,
    n.name AS neighborhood_name,
    n.lat AS neighborhood_lat,
    n.lon AS neighborhood_lon,
    c.name AS city_name,
    c.lat AS city_lat,
    c.lon AS city_lon,
    st.uf AS state_uf,
    st.lat AS state_lat,
    st.lon AS state_lon,
    a.id AS address_id,
    a.number,
    a.cep,
    a.lat,
    a.lon
FROM addresses a
         JOIN streets s ON a.street_id = s.id
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
WHERE a.cep = $1
limit 5;

-- name: FindStatesByName :many
SELECT *
FROM states
WHERE search_vector @@ plainto_tsquery('portuguese', $1);

-- name: FindCitiesByName :many
SELECT c.*, s.uf
FROM cities c
JOIN states s ON c.state_id = s.id
WHERE c.search_vector @@ plainto_tsquery('portuguese', $1);

-- name: FindNeighborhoodsByName :many
SELECT n.*, c.name as city, s.uf
FROM neighborhoods n
JOIN cities c ON n.city_id = c.id
JOIN states s ON c.state_id = s.id
WHERE n.search_vector @@ plainto_tsquery('portuguese', $1);

-- name: FindStateAll :many
select * from states order by name;

-- name: FindCityAll :many
select * from cities c where c.state_id = $1 order by name;

-- name: FindAddressGroupedByCEP :one
SELECT
    city_name,
    state_uf,
    neighborhood_name,
    street_name,
    lat AS latitude,
    lon AS longitude
FROM public.unique_ceps
WHERE cep = $1 LIMIT 1;

-- name: FindTwoRandomCEPs :many
SELECT
    cep::text AS cep
FROM public.unique_ceps
ORDER BY RANDOM()
    LIMIT 2;

-- name: FindAddressByStreetID :many
SELECT * FROM addresses a
WHERE a.street_id = $1;

-- name: FindAddressByCEP :one
SELECT
    c.name AS city_name,
    st.uf AS state_uf,
    n.name as neighborhood_name,
    s.name as street_name,
    a.lat as latitude,
    a.lon as longitude
FROM addresses a
         JOIN streets s ON a.street_id = s.id
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
WHERE a.cep = $1
GROUP BY c.name, st.uf,  n.name, s.name, a.lat, a.lon
    LIMIT 1;

-- name: FindAddressByCEPNew :one
select id, lat, lon from addresses_coordinates a where a.cep = $1 limit 1;

-- name: FindUniqueAddressByCEP :many
SELECT
    u.street_id,
    u.street_name,
    u.neighborhood_name,
    u.neighborhood_lat,
    u.neighborhood_lon,
    u.city_name,
    u.city_lat,
    u.city_lon,
    u.state_uf,
    u.state_lat,
    u.state_lon,
    u.id as address_id,
    u.number,
    u.cep,
    u.lat,
    u.lon
FROM unique_ceps u
WHERE u.cep = $1
    limit 5;