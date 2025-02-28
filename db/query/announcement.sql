-- name: CreateAnnouncement :one
INSERT INTO public.announcement
(id, user_id, destination, origin, destination_lat, destination_lng, origin_lat, origin_lng, distance, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_volume, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, price, status, created_at, created_who)
VALUES(nextval('announcement_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9,
       $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, true, now(), $28)
    RETURNING *;

-- name: UpdateAnnouncement :one
UPDATE public.announcement
SET user_id=$1, destination=$2, origin=$3, destination_lat=$4, destination_lng=$5, origin_lat=$6, origin_lng=$7, distance=$8, pickup_date=$9, delivery_date=$10, expiration_date=$11, title=$12,
    cargo_type=$13, cargo_species=$14, cargo_volume=$15, cargo_weight=$16, vehicles_accepted=$17, trailer=$18, requires_tarp=$19, tracking=$20, agency=$21, description=$22, payment_type=$23, advance=$24, toll=$25, situation=$26, price=$27, updated_at=now(), updated_who=$28
WHERE id=$29
    RETURNING *;

-- name: DeleteAnnouncement :exec
UPDATE public.announcement
SET status=false, updated_at=now(), updated_who=$2
WHERE id=$1;

-- name: GetAnnouncementById :one
SELECT *
FROM public.announcement
WHERE id=$1;


-- name: GetAllAnnouncementUsers :one
SELECT *
FROM public.announcement
WHERE status=true
ORDER BY expiration_date;

-- name: GetAllAnnouncementPublic :one
SELECT id, destination, origin, pickup_date, delivery_date, expiration_date, title, cargo_type, cargo_species, cargo_volume, cargo_weight, vehicles_accepted, trailer, requires_tarp, tracking, agency, description, payment_type, advance, toll, situation, created_at
FROM public.announcement
WHERE status=true
ORDER BY expiration_date;