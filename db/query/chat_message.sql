-- name: CreateChatMessage :one
INSERT INTO chat_messages
(room_id, user_id, content, status, is_read)
VALUES
($1, $2, $3, true, false)
RETURNING *;

