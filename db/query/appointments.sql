-- name: CreateAppointment :one
INSERT INTO appointments
(id, advertisement_user_id, interested_user_id, offer_id, truck_id, advertisement_id, situation, status, created_who, created_at)
VALUES(nextval('appointments_id_seq'::regclass), $1, $2, $3, $4, $5,'ATIVO',
       true,$6,now())
    RETURNING *;

-- name: UpdateAppointmentSituation :exec
UPDATE appointments
SET situation=$1, updated_who=$2, updated_at=now()
WHERE id=$3
    RETURNING *;

-- name: DeleteAppointment :exec
UPDATE appointments
SET status=false, updated_at=now()
WHERE id=$1;

-- name: GetAppointmentByID :one
SELECT *
FROM appointments
WHERE id=$1;

-- name: GetListAppointmentByUserID :many
SELECT DISTINCT ON (ap.id) *
FROM appointments ap
         LEFT JOIN users u ON u.id IN (ap.advertisement_user_id, ap.interested_user_id) AND u.status = true
         LEFT JOIN advertisement ad ON ad.id = ap.advertisement_id AND ad.status = true
         LEFT JOIN truck tr ON tr.id = ap.truck_id
         LEFT JOIN tractor_unit tu ON tu.id = tr.tractor_unit_id AND tu.status = true
         LEFT JOIN trailer t ON t.id = tr.trailer_id AND t.status = true
         LEFT JOIN driver d ON d.id = tr.driver_id AND d.status = true
WHERE (ap.advertisement_user_id=$1 OR ap.interested_user_id=$1) AND ap.status = true;


