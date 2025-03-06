-- name: CreateChatMessage :one
INSERT INTO chat_messages
(room_id, user_id, content, status, is_read, type_message)
VALUES
($1, $2, $3, true, false, $4)
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

-- name: UpdateMessageStatus :exec
UPDATE chat_messages
SET is_accepted = $1
WHERE id = $2;

-- name: GetRoomByMessageId :one
SELECT r.*, m.type_message, m.is_accepted FROM chat_messages m
JOIN chat_rooms r ON r.id = m.room_id
WHERE m.id = $1;