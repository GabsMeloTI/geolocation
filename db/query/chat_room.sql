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
SELECT 
    r.id AS room_id, 
    r.created_at, 
    r.advertisement_user_id, 
    a.id AS advertisement_id, 
    a.origin, 
    a.destination, 
    a.distance, 
    a.title,
    (SELECT COUNT(m.is_read) 
     FROM chat_messages m
     WHERE m.room_id = r.id and m.is_read = FALSE and m.user_id != $1) AS unread_count
FROM chat_rooms r
JOIN "advertisement" a ON a.id = r.advertisement_id
WHERE r.interested_user_id = $1;

-- name: GetAdvertisementChatRooms :many
SELECT 
    r.id AS room_id, 
    r.created_at, 
    r.advertisement_user_id, 
    a.id AS advertisement_id, 
    a.origin, 
    a.destination, 
    a.distance, 
    a.title,
  COUNT(m.is_read) FILTER(WHERE m.is_read = FALSE AND m.user_id !=$1) as unread_count
    FROM chat_rooms r
JOIN "chat_messages" m ON m.room_id = r.id
JOIN "advertisement" a ON a.id = r.advertisement_id
WHERE r.advertisement_user_id = $1
AND EXISTS (SELECT 1 FROM chat_messages cm WHERE cm.room_id = r.id)
GROUP BY r.id, a.id;

-- name: GetChatRoomByAdvertisementAndInterestedUser :one
select r.* from chat_rooms r
where r.advertisement_id = $1 and r.interested_user_id = $2 AND status;
