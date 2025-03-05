-- name: CreateUserPlans :one
INSERT INTO public.user_plan
(id, id_user, id_plan, annual, active, active_date, expiration_date)
VALUES(nextval('user_plan_id_seq'::regclass), $1, $2, $3, true, now(), $4)
    RETURNING *;


-- name: GetUserPlanByIdUser :one
SELECT *
FROM public.user_plan
WHERE active=true AND
      id_user=$1 AND
      id_plan=$2;

-- name: UpdateUserPlan :exec
UPDATE public.user_plan
SET active=false
WHERE id_user=$1 AND
      id_plan=$2;