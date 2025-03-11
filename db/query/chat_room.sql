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

-- name: GetInterestedChatRooms :many
SELECT r.id as room_id, r.created_at, r.advertisement_user_id, a.id as advertisement_id, a.origin, a.destination, a.distance, a.title FROM chat_rooms r
JOIN "advertisement" a ON a.id = r.advertisement_id
WHERE r.interested_user_id = $1;


-- name: GetAdvertisementChatRooms :many
SELECT r.id as room_id, r.created_at, r.advertisement_user_id, a.id as advertisement_id, a.origin, a.destination, a.distance, a.title FROM chat_rooms r
JOIN "advertisement" a ON a.id = r.advertisement_id
WHERE r.advertisement_user_id = $1;


-- name: GetChatRoomByAdvertisementAndInterestedUser :one
select r.* from chat_rooms r
where r.advertisement_id = $1 and r.interested_user_id = $2 AND status;