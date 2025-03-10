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
    a.lon
FROM streets s
         LEFT JOIN neighborhoods n ON s.neighborhood_id = n.id
         JOIN cities c ON n.city_id = c.id
         JOIN states st ON c.state_id = st.id
         LEFT JOIN addresses a ON a.street_id = s.id
WHERE
    (COALESCE(NULLIF($1, ''), NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, '')) IS NOT NULL)
  AND (s.search_vector @@ plainto_tsquery('portuguese', $1) OR $1 = '')
  AND (c.search_vector @@ plainto_tsquery('portuguese', $2) OR $2 = '')
  AND (st.search_vector @@ plainto_tsquery('portuguese', $3) OR $3 = '')
  AND (n.search_vector @@ plainto_tsquery('portuguese', $4) OR $4 = '')
  AND ($5 = '' OR a.number ILIKE $5 || '%')
ORDER BY random();

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
ORDER BY (a.lat - $1) * (a.lat - $1) + (a.lon - $2) * (a.lon - $2) ASC;

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
WHERE a.cep = $1;

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