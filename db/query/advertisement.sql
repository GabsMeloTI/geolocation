-- name: CreateAdvertisement :one
INSERT INTO public.advertisement
(id, user_id, destination, origin, destination_lat, destination_lng, origin_lat, origin_lng, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, price, status, created_at, created_who, state, city, complement, neighborhood, street, street_number, cep)
VALUES(nextval('advertisement_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9,
       $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26,  true, now(), $27, $28, $29, $30, $31, $32, $33, $34)
    RETURNING *;

-- name: UpdateAdvertisement :one
UPDATE public.advertisement
SET user_id=$1, destination=$2, origin=$3, destination_lat=$4, destination_lng=$5, origin_lat=$6, origin_lng=$7, distance=$8, pickup_date=$9, delivery_date=$10, expiration_date=$11, title=$12,
    cargo_type=$13, cargo_species=$14, cargo_weight=$15, vehicles_accepted=$16, trailer=$17, requires_tarp=$18, tracking=$19, agency=$20, description=$21, payment_type=$22, advance=$23, toll=$24, situation=$25, price=$26, updated_at=now(), updated_who=$27,
    state=$28, city=$29, complement=$30, neighborhood=$31, street=$32, street_number=$33, cep=$34
WHERE id=$35
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
SELECT a.id, user_id, u.name as user_name, u.created_at as active_there, u.city as user_city, u.state as user_state, u.phone as user_phone, u.email as user_email, u.profile_picture as user_profile_picture, destination, origin, destination_lat, destination_lng, origin_lat, origin_lng, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, price, a.created_at, created_who, a.updated_at, updated_who,
       a.state, a.city, a.complement, a.neighborhood, a.street, a.street_number, a.cep
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
SELECT id, destination, origin, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, created_at,
       state, city, complement, neighborhood, street, street_number, cep
FROM public.advertisement
WHERE status=true
ORDER BY expiration_date;
