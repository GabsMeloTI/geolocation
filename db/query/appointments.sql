-- name: CreateAppointment :one
INSERT INTO appointments
(id, user_id, truck_id, advertisement_id, situation, status, created_who, created_at)
VALUES(nextval('appointments_id_seq'::regclass), $1, $2, $3, 'ATIVO',
       true,$4,now())
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
SELECT *
FROM appointments ap
    LEFT JOIN users u ON u.id = ap.user_id AND u.status = true
    LEFT JOIN advertisement ad ON ad.id = ap.advertisement_id AND ad.status = true
    LEFT JOIN truck tr ON tr.id = ap.truck_id
    LEFT JOIN tractor_unit tu ON tu.id = tr.tractor_unit_id AND tu.status = true
    LEFT JOIN trailer t ON t.id = tr.trailer_id AND t.status = true
    LEFT JOIN driver d ON d.id = tr.driver_id AND d.status = true
WHERE ap.user_id=$1 and ap.status=true;

-- name: GetListAppointmentByAdvertiser :many
SELECT *
FROM appointments ap
         LEFT JOIN users u ON u.id = ap.user_id AND u.status = true
         LEFT JOIN advertisement ad ON ad.id = ap.advertisement_id AND ad.status = true
         LEFT JOIN truck tr ON tr.id = ap.truck_id
         LEFT JOIN tractor_unit tu ON tu.id = tr.tractor_unit_id AND tu.status = true
         LEFT JOIN trailer t ON t.id = tr.trailer_id AND t.status = true
         LEFT JOIN driver d ON d.id = tr.driver_id AND d.status = true
WHERE ap.advertisement_id in (SELECT ad.id FROM advertisement ad where ad.user_id = $1)
  and ap.status=true;