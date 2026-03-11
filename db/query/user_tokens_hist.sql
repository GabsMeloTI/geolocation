-- name: CreateUserTokenHist :one
INSERT INTO user_tokens_hist (user_id, token, payload, expired_at, origin)
VALUES ($1, $2, $3, $4, $5) RETURNING *;
