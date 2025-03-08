-- name: CreatePaymentHist :one
INSERT INTO public.payment_hist
(id, user_id, email, "name", value, "method", automatic, payment_date, payment_expireted, payment_status, currency, invoice, customer, "interval")
VALUES(nextval('payment_hist_id_seq'::regclass), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    RETURNING *;

-- name: GetPaymentHist :many
SELECT *
FROM public.payment_hist
WHERE user_id=$1;
