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
    (COALESCE(NULLIF(sqlc.arg(street), ''), NULLIF(sqlc.arg(city), ''), NULLIF(sqlc.arg(state), ''), NULLIF(sqlc.arg(neighborhood), '')) IS NOT NULL)
  AND (s.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(street)) OR sqlc.arg(street) = '')
  AND (c.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(city)) OR sqlc.arg(city) = '')
  AND (st.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(state)) OR sqlc.arg(state) = '')
  AND (n.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(neighborhood)) OR sqlc.arg(neighborhood) = '')
  AND (sqlc.arg(number) = '' OR a.number ILIKE sqlc.arg(number) || '%')
ORDER BY random()
limit 100;

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
limit 100;

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
limit 100;

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

-- name: FindAddressGroupedByCEP :many
SELECT
    c.name AS city_name,
    st.uf AS state_uf,
    n.name as neighborhood_name,
    s.name as street_name,
    a.lat as latitude,
    a.lat as longitude
FROM addresses a
         JOIN streets s ON a.street_id = s.id
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
WHERE a.cep = $1
GROUP BY c.name, st.uf,  n.name, s.name, a.lat, a.lon
    LIMIT 100;

-- name: FindAddressByCEP :one
SELECT
    c.name AS city_name,
    st.uf AS state_uf,
    n.name as neighborhood_name,
    s.name as street_name,
    a.lat as latitude,
    a.lat as longitude
FROM addresses a
         JOIN streets s ON a.street_id = s.id
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
WHERE a.cep = $1
GROUP BY c.name, st.uf,  n.name, s.name, a.lat, a.lon
    LIMIT 1;