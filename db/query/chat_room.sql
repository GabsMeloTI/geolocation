-- name: CreateChatRoom :one

INSERT INTO chat_rooms
(id, advertisement_id, advertisement_user_id, interested_user_id, status, created_at)
VALUES
(nextval('chat_rooms_id_seq'::regclass), $1, $2, $3, true, now())
RETURNING *;

-- name: GetChatRoomById :one
SELECT r.*, a.id as advertisement_id, a.origin, a.destination, a.distance FROM chat_rooms r
JOIN "advertisement" a ON a.id = r.advertisement_id
WHERE r.id = $1;

-- name: GetDriverChatRooms :many
SELECT r.*, a.id as advertisement_id, a.origin, a.destination, a.distance FROM chat_rooms r
JOIN "advertisement" a ON a.id = r.advertisement_id
WHERE r.interested_user_id = $1;


-- name: GetCarrierChatRooms :many
=SELECT r.*, a.id as advertisement_id, a.origin, a.destination, a.distance FROM chat_rooms r
JOIN "advertisement" a ON a.id = r.advertisement_id
WHERE r.interested_user_id = $1;