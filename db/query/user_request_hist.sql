-- name: CreateUserRequestHist :one
INSERT INTO user_request_hist (user_id, token, endpoint, method, status_code, execution_time_ms)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
