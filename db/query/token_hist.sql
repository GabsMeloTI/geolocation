-- name: CreateTokenHist :one
INSERT INTO public.token_hist
(id, ip, number_request, created_at, exprited_at, "valid")
VALUES(nextval('token_hist_id_seq'::regclass), $1, $2, now(), $3, true)
    RETURNING *;

-- name: UpdateNumberOfRequest :exec
UPDATE public.token_hist
SET number_request = $1
WHERE id = $2;

-- name: GetTokenHist :one
SELECT *
FROM public.token_hist
WHERE id=$1;