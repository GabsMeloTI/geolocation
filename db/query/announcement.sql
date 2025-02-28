-- name: CreateAdvertisement :one
INSERT INTO public.advertisement
(id, user_id, destination, origin, destination_lat, destination_lng, origin_lat, origin_lng, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_volume, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, price, status, created_at, created_who)
VALUES(nextval('advertisement_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9,
       $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, true, now(), $28)
    RETURNING *;

-- name: UpdateAdvertisement :one
UPDATE public.advertisement
SET user_id=$1, destination=$2, origin=$3, destination_lat=$4, destination_lng=$5, origin_lat=$6, origin_lng=$7, distance=$8, pickup_date=$9, delivery_date=$10, expiration_date=$11, title=$12,
    cargo_type=$13, cargo_species=$14, cargo_volume=$15, cargo_weight=$16, vehicles_accepted=$17, trailer=$18, requires_tarp=$19, tracking=$20, agency=$21, description=$22, payment_type=$23, advance=$24, toll=$25, situation=$26, price=$27, updated_at=now(), updated_who=$28
WHERE id=$29
    RETURNING *;

-- name: DeleteAdvertisement :exec
UPDATE public.advertisement
SET status=false, updated_at=now(), updated_who=$2
WHERE id=$1;

-- name: GetAdvertisementById :one
SELECT *
FROM public.advertisement
WHERE id=$1;


-- name: GetAllAdvertisementUsers :many
SELECT a.id, user_id, u.name as user_name, u.created_at as active_there, u.city as user_city, u.state as user_state, u.phone as user_phone, u.email as user_email, u.profile_picture as user_profile_picture, destination, origin, destination_lat, destination_lng, origin_lat, origin_lng, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_volume, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, a.created_at, created_who, a.updated_at, updated_who
FROM public.advertisement a
         inner join users u on u.id = a.user_id
WHERE a.status=true
ORDER BY expiration_date;


-- name: CountAdvertisementByUserID :one
SELECT COUNT(*)
FROM public.advertisement
WHERE user_id = $1
  AND status = true
  AND situation = 'ativo';


-- name: GetAllAdvertisementPublic :many
SELECT id, destination, origin, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_volume, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, created_at
FROM public.advertisement
WHERE status=true
ORDER BY expiration_date;