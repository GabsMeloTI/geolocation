-- name: CreateOffer :one
INSERT INTO offers
(advertisement_id, price, interested_id)
VALUES
($1, $2, $3)
RETURNING *;
