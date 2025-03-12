-- name: CreateAdvertisement :one
INSERT INTO public.advertisement
(id, user_id, destination, origin, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_weight, vehicles_accepted, trailer,
 requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, price, status, created_at, created_who, state_origin, city_origin, complement_origin, neighborhood_origin, street_origin, street_number_origin,
 cep_origin, state_destination, city_destination, complement_destination, neighborhood_destination, street_destination, street_number_destination, cep_destination)
VALUES(nextval('advertisement_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9,
       $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, 'pendente', $21,
       true, now(),$22,$23, $24, $25, $26, $27,$28, $29, $30, $31,
       $32, $33, $34,$35, $36)
    RETURNING *;

-- name: UpdatedAdvertisementFinishedCreate :one
UPDATE public.advertisement
SET destination_lat=$3, destination_lng=$4, origin_lat=$5, origin_lng=$6, situation='ativo'
WHERE user_id=$1 AND
      id=$2
      RETURNING id, user_id, destination_lat, destination_lng, origin_lat, origin_lng, situation;


-- name: UpdateAdvertisement :one
UPDATE public.advertisement
SET destination=$2, origin=$3, destination_lat=$4, destination_lng=$5, origin_lat=$6, origin_lng=$7, distance=$8, pickup_date=$9, delivery_date=$10, expiration_date=$11, title=$12,
    cargo_type=$13, cargo_species=$14, cargo_weight=$15, vehicles_accepted=$16, trailer=$17, requires_tarp=$18, tracking=$19, agency=$20, description=$21, payment_type=$22, advance=$23, toll=$24, situation=$25, price=$26, updated_at=now(), updated_who=$27,
    state_origin=$28, city_origin=$29, complement_origin=$30, neighborhood_origin=$31, street_origin=$32, street_number_origin=$33, cep_origin=$34,
    state_destination=$35, city_destination=$36, complement_destination=$37, neighborhood_destination=$38, street_destination=$39, street_number_destination=$40, cep_destination=$41
WHERE user_id=$1 AND
    id=$42
    RETURNING *;


-- name: DeleteAdvertisement :exec
UPDATE public.advertisement
SET status=false, updated_at=now(), updated_who=$3
WHERE id=$1 and
    user_id=$2;

-- name: GetAdvertisementById :one
SELECT *
FROM public.advertisement
WHERE id=$1 AND
    status=true;

-- name: GetAllAdvertisementUsers :many
SELECT a.id, a.user_id, u.name as user_name, u.created_at as active_there, u.city as user_city, u.state as user_state, u.phone as user_phone, u.email as user_email, u.profile_picture as user_profile_picture,
       a.destination, a.origin, destination_lat, destination_lng, origin_lat, origin_lng, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, price, a.created_at, created_who, a.updated_at, updated_who,
       a.state_origin, a.city_origin, a.complement_origin, a.neighborhood_origin, a.street_origin, a.street_number_origin, a.cep_origin,
       a.state_destination, a.city_destination, a.complement_destination, a.neighborhood_destination, a.street_destination, a.street_number_destination, a.cep_destination, rh.response as response_routes, ar.route_choose
FROM public.advertisement a
         INNER JOIN users u ON u.id = a.user_id
         INNER JOIN advertisement_route ar on a.id = ar.advertisement_id
         INNER JOIN route_hist rh on rh.id = ar.route_hist_id
WHERE a.status=true AND destination_lat IS NOT NULL AND destination_lng IS NOT NULL AND origin_lat IS NOT NULL AND origin_lng IS NOT NULL
ORDER BY expiration_date;

-- name: CountAdvertisementByUserID :one
SELECT COUNT(*)
FROM public.advertisement
WHERE user_id = $1
  AND status = true AND destination_lat IS NOT NULL AND destination_lng IS NOT NULL AND origin_lat IS NOT NULL AND origin_lng IS NOT NULL AND situation = 'ativo';


-- name: GetAllAdvertisementPublic :many
SELECT id, user_id, destination, origin, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, created_at,
       state_origin, city_origin, complement_origin, neighborhood_origin, street_origin, street_number_origin, cep_origin,
       state_destination, city_destination, complement_destination, neighborhood_destination, street_destination, street_number_destination, cep_destination
FROM public.advertisement
WHERE status=true AND destination_lat IS NOT NULL AND destination_lng IS NOT NULL AND origin_lat IS NOT NULL AND origin_lng IS NOT NULL
ORDER BY expiration_date;

-- name: UpdateAdvertisementSituation :exec
UPDATE public.advertisement
SET situation=$1, updated_at=now(), updated_who=$2
WHERE id=$3;

-- name: CreateAdvertisementRoute :one
INSERT INTO public.advertisement_route
(id, advertisement_id, route_hist_id, user_id, route_choose, created_at)
VALUES(nextval('advertisement_route_id_seq'::regclass), $1, $2, $3, $4, now())
    RETURNING *;


-- name: GetAllAdvertisementByUser :many
SELECT a.id, a.user_id, u.name as user_name, u.created_at as active_there, u.city as user_city, u.state as user_state, u.phone as user_phone, u.email as user_email, u.profile_picture as user_profile_picture,
       a.destination, a.origin, destination_lat, destination_lng, origin_lat, origin_lng, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, price, a.created_at, created_who, a.updated_at, updated_who,
       a.state_origin, a.city_origin, a.complement_origin, a.neighborhood_origin, a.street_origin, a.street_number_origin, a.cep_origin,
       a.state_destination, a.city_destination, a.complement_destination, a.neighborhood_destination, a.street_destination, a.street_number_destination, a.cep_destination, rh.response as response_routes, ar.route_choose
FROM public.advertisement a
         INNER JOIN users u ON u.id = a.user_id
         LEFT JOIN advertisement_route ar on a.id = ar.advertisement_id
         LEFT JOIN route_hist rh on rh.id = ar.route_hist_id
WHERE a.status=true AND a.user_id=$1
ORDER BY expiration_date;

-- name: UpdateAdsRouteChooseByUserId :exec
UPDATE advertisement_route SET route_choose = $1 WHERE user_id = $2 AND advertisement_id = $3;

