-- name: CreateChatMessage :one
INSERT INTO chat_messages
(room_id, user_id, content, status, is_read)
VALUES
($1, $2, $3, true, false)
RETURNING *;


-- name: GetChatMessagesByRoomId :many
SELECT m.*, u.name, u.profile_picture
FROM public.chat_messages m JOIN users u on m.user_id = u.id
JOIN chat_rooms r on r.id = m.room_id AND (r.advertisement_user_id = @user_id OR r.interested_user_id = @user_id)
WHERE m.room_id = $1
ORDER BY m.created_at ASC;


-- name: GetLastMessageByRoomId :many
SELECT m.*, u.name, u.profile_picture from chat_messages m
JOIN users u on m.user_id = u.id
JOIN chat_rooms r on r.id = m.room_id AND (r.advertisement_user_id = @user_id OR r.interested_user_id = @user_id)
ORDER BY m.created_at DESC
LIMIT 1;