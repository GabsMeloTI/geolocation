-- name: CreateTokenHist :one
INSERT INTO public.token_hist
(id, ip, number_request, created_at, exprited_at, "valid", token)
VALUES(nextval('token_hist_id_seq'::regclass), $1, $2, now(), $3, true, $4)
    RETURNING *;

-- name: UpdateNumberOfRequest :exec
UPDATE public.route_hist
SET number_request = $1
WHERE id = $2;

-- name: GetTokenHist :one
SELECT *
FROM public.token_hist
WHERE ip=$1;

-- name: GetTokenHistExist :one
SELECT EXISTS (
    SELECT 1
    FROM public.token_hist
    WHERE ip = $1
);

-- name: UpdateTokenHist :one
UPDATE public.token_hist
SET number_request = $2,
    exprited_at = $3,
    token = $4,
    created_at = now()
WHERE id = $1
    RETURNING id, ip, number_request, valid, exprited_at, token;
