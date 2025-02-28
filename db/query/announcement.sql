-- name: CreateAnnouncement :one
INSERT INTO public.announcement
(id, destination, origin, destination_lat, destination_lng, origin_lat, origin_lng, description, cargo_description, payment_description, delivery_date, pickup_date, deadline_date, price, vehicle, body_type, kilometers, cargo_nature, cargo_type, cargo_weight, tracking, requires_tarp, status, created_at)
VALUES(nextval('announcement_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9,
       $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, true, now())
    RETURNING *;

-- name: UpdateAnnouncement :one
UPDATE public.announcement
SET destination=$2, origin=$3, destination_lat=$4, destination_lng=$5, origin_lat=$6, origin_lng=$7, description=$8, cargo_description=$9, payment_description=$10, delivery_date=$11, pickup_date=$12,
    deadline_date=$13, price=$14, vehicle=$15, body_type=$16, kilometers=$17, cargo_nature=$18, cargo_type=$19, cargo_weight=$20, tracking=$21, requires_tarp=$22, updated_at=now()
WHERE id=$1
    RETURNING *;

-- name: DeleteAnnouncement :exec
UPDATE public.announcement
SET status=false, updated_at=now()
WHERE id=$1;

-- name: GetAnnouncementById :one
SELECT *
FROM public.announcement
WHERE id=$1;
