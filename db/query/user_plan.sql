-- name: CreateUserPlans :one
INSERT INTO public.user_plan
(id, id_user, id_plan, active, active_date, expiration_date)
VALUES(nextval('user_plan_id_seq'::regclass), $1, $2, true, now(), $3)
    RETURNING *;


