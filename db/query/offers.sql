-- name: CreateOffer :one
INSERT INTO offers
(advertisement_id, price, interested_id, status)
VALUES
($1, $2, $3 , $4)
RETURNING *;
