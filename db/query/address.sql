-- name: FindAddressesByQuery :many
SELECT
    s.id AS street_id,
    s.name AS street_name,
    n.name AS neighborhood_name,
    c.name AS city_name,
    st.uf AS state_uf,
    a.id AS address_id,
    a.number,
    a.cep,
    a.lat,
    a.lon,
    ts_rank(s.search_vector, plainto_tsquery('portuguese', sqlc.arg(street))) AS street_rank,
    ts_rank(c.search_vector, plainto_tsquery('portuguese', sqlc.arg(city))) AS city_rank,
    ts_rank(st.search_vector, plainto_tsquery('portuguese', sqlc.arg(state))) AS state_rank,
    ts_rank(n.search_vector, plainto_tsquery('portuguese', sqlc.arg(neighborhood))) AS neighborhood_rank
FROM streets s
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
         LEFT JOIN addresses a ON a.street_id = s.id
WHERE
    (COALESCE(NULLIF(sqlc.arg(street), ''),
              NULLIF(sqlc.arg(city), ''),
              NULLIF(sqlc.arg(state), ''),
              NULLIF(sqlc.arg(neighborhood), '')) IS NOT NULL)
  AND (s.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(street)) OR sqlc.arg(street) = '')
  AND (c.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(city)) OR sqlc.arg(city) = '')
  AND (st.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(state)) OR sqlc.arg(state) = '')
  AND (n.search_vector @@ plainto_tsquery('portuguese', sqlc.arg(neighborhood)) OR sqlc.arg(neighborhood) = '')
  AND (sqlc.arg(number) = '' OR a.number ILIKE sqlc.arg(number) || '%')
ORDER BY
    ts_rank(s.search_vector, plainto_tsquery('portuguese', sqlc.arg(street))) +
    ts_rank(c.search_vector, plainto_tsquery('portuguese', sqlc.arg(city))) +
    ts_rank(st.search_vector, plainto_tsquery('portuguese', sqlc.arg(state))) +
    ts_rank(n.search_vector, plainto_tsquery('portuguese', sqlc.arg(neighborhood))) DESC
LIMIT 5;

-- name: FindAddressesByLatLon :many
SELECT
    a.id AS address_id,
    s.name AS street_name,
    s.id AS street_id,
    n.name AS neighborhood_name,
    c.name AS city_name,
    st.uf AS state_uf,
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
    a.id AS address_id,
    s.name AS street_name,
    s.id AS street_id,
    n.name AS neighborhood_name,
    c.name AS city_name,
    st.uf AS state_uf,
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

-- name: IsState :one
SELECT EXISTS(
    SELECT 1
    FROM states
    WHERE search_vector @@ plainto_tsquery('portuguese', $1)
) AS is_state;

-- name: IsCity :one
SELECT EXISTS(
    SELECT 1
    FROM cities
    WHERE search_vector @@ plainto_tsquery('portuguese', $1)
) AS is_city;

-- name: IsNeighborhood :one
SELECT EXISTS(
    SELECT 1
    FROM neighborhoods
    WHERE search_vector @@ plainto_tsquery('portuguese', $1)
) AS is_neighborhood;

-- name: FindStateAll :many
select * from states order by name;

-- name: FindCityAll :many
select * from cities c where c.state_id = $1 order by name;

-- name: FindAddressGroupedByCEP :many
SELECT
    c.name AS city_name,
    st.uf AS state_uf,
    n.name as neighborhood_name,
    s.name as street_name
FROM addresses a
         JOIN streets s ON a.street_id = s.id
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
WHERE a.cep = $1
GROUP BY c.name, st.uf,  n.name, s.name
LIMIT 5;