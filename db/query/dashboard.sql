-- name: GetDashboardDriver :one
WITH faturamento_mensal AS (
    SELECT
        ap.interested_user_id,
        EXTRACT(YEAR FROM a.delivery_date) AS ano,
        EXTRACT(MONTH FROM a.delivery_date) AS mes,
        SUM(a.price::float8) AS total_faturado
    FROM appointments ap
             INNER JOIN advertisement a ON a.id = ap.advertisement_id
    WHERE ap.situation = 'finalizado'
    GROUP BY ap.interested_user_id,
             EXTRACT(YEAR FROM a.delivery_date),
             EXTRACT(MONTH FROM a.delivery_date)
)
SELECT
    u.id AS user_id,
    d.id AS driver_id,
    COALESCE(
            SUM(
                    CASE
                        WHEN ap.situation = 'finalizado'
                            AND EXTRACT(MONTH FROM a.delivery_date) = EXTRACT(MONTH FROM CURRENT_DATE)
                            AND EXTRACT(YEAR FROM a.delivery_date) = EXTRACT(YEAR FROM CURRENT_DATE)
                            THEN a.price::float8
                        ELSE 0::float8
                        END
            ), 0::float8
    )::float8 AS total_fretes_finalizados_mes_atual,
    COALESCE(
            SUM(
                    CASE
                        WHEN ap.situation != 'finalizado'
                        AND EXTRACT(MONTH FROM a.delivery_date) = EXTRACT(MONTH FROM CURRENT_DATE)
                        AND EXTRACT(YEAR FROM a.delivery_date) = EXTRACT(YEAR FROM CURRENT_DATE)
                    THEN a.price::float8
                    ELSE 0::float8
                    END
            ), 0::float8
    )::float8 AS total_a_receber_mes_atual,
    COUNT(
            DISTINCT CASE
                         WHEN ap.situation = 'finalizado'
                             AND EXTRACT(MONTH FROM a.delivery_date) = EXTRACT(MONTH FROM CURRENT_DATE)
                             AND EXTRACT(YEAR FROM a.delivery_date) = EXTRACT(YEAR FROM CURRENT_DATE)
                             THEN ap.advertisement_user_id
        END
    ) AS clientes_atendidos_mes_atual,
    COALESCE(
            SUM(
                    CASE
                        WHEN ap.situation = 'finalizado'
                            AND EXTRACT(MONTH FROM a.delivery_date) = CASE WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 THEN 12 ELSE EXTRACT(MONTH FROM CURRENT_DATE) - 1 END
                            AND EXTRACT(YEAR FROM a.delivery_date) = CASE WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 THEN EXTRACT(YEAR FROM CURRENT_DATE) - 1 ELSE EXTRACT(YEAR FROM CURRENT_DATE) END
                            THEN a.price::float8
                        ELSE 0::float8
                        END
            ), 0::float8
    )::float8 AS total_fretes_finalizados_mes_anterior,
    COUNT(
            DISTINCT CASE
                         WHEN ap.situation = 'finalizado'
                             AND EXTRACT(MONTH FROM a.delivery_date) = CASE WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 THEN 12 ELSE EXTRACT(MONTH FROM CURRENT_DATE) - 1 END
                             AND EXTRACT(YEAR FROM a.delivery_date) = CASE WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 THEN EXTRACT(YEAR FROM CURRENT_DATE) - 1 ELSE EXTRACT(YEAR FROM CURRENT_DATE) END
                             THEN ap.advertisement_user_id
        END
    ) AS clientes_atendidos_mes_anterior
FROM users u
         INNER JOIN driver d ON d.user_id = u.id
         INNER JOIN appointments ap ON ap.interested_user_id = u.id
         INNER JOIN advertisement a ON a.id = ap.advertisement_id
         LEFT JOIN faturamento_mensal fm ON fm.interested_user_id = u.id
WHERE u.id = $1
GROUP BY u.id, d.id;



-- name: GetDashboardHist :many
SELECT
    adv.name AS historico_user_name,
    a.id AS historico_frete_id,
    a.delivery_date AS historico_delivery_date,
    a.price AS historico_price
FROM users u
         INNER JOIN driver d ON d.user_id = u.id
         INNER JOIN appointments ap ON ap.interested_user_id = u.id
         INNER JOIN advertisement a ON a.id = ap.advertisement_id
         INNER JOIN users adv ON adv.id = ap.advertisement_user_id
WHERE u.id = $1 AND ap.situation = 'finalizado'
GROUP BY adv.name, a.id, a.delivery_date, a.price;



-- name: GetDashboardFuture :many
SELECT
    a.user_id,
    a.id AS advertisement_id,
    a.origin AS lembretes_origin,
    a.destination AS lembretes_destination,
    a.pickup_date AS lembretes_pickup_date
FROM advertisement a
WHERE a.user_id = $1 AND a.pickup_date > NOW()
GROUP BY a.id, a.user_id, a.origin, a.destination, a.pickup_date
ORDER BY a.pickup_date ASC;


-- name: GetDashboardCalendar :many
SELECT
    a.user_id,
    a.id AS advertisement_id,
    ap.id AS appointments_id,
    a.pickup_date,
    ap.situation
FROM appointments ap
         INNER JOIN advertisement a ON a.id = ap.advertisement_id
WHERE a.user_id = $1
ORDER BY a.pickup_date ASC;


-- name: GetDashboardFaturamento :many
WITH faturamento_mensal AS (
    SELECT
        ap.interested_user_id,
        EXTRACT(YEAR FROM ap.created_at)::bigint AS ano,
            EXTRACT(MONTH FROM ap.created_at)::bigint AS mes,
            SUM(a.price)::float8 AS total_faturado
    FROM appointments ap
             INNER JOIN advertisement a ON a.id = ap.advertisement_id
    WHERE ap.situation = 'finalizado'
    GROUP BY ap.interested_user_id,
             EXTRACT(YEAR FROM ap.created_at),
             EXTRACT(MONTH FROM ap.created_at)
)
SELECT
    ano,
    mes,
    total_faturado
FROM faturamento_mensal
WHERE interested_user_id = $1
ORDER BY ano DESC, mes DESC;



-- name: GetDashboardDriverEnterprise :many
select d.id, d.name, d.license_number,  d.license_category,
       CASE
           WHEN EXISTS (
               SELECT 1
               FROM truck t
                        JOIN appointments a ON t.id = a.truck_id
               WHERE t.driver_id = d.id
                 AND a.situation = 'em andamento'
           ) THEN 'em viagem'
           ELSE 'disponivel'
           END AS disponibilidade
FROM driver d
WHERE d.user_id = $1;

-- name: GetDashboardTractUnitEnterprise :many
SELECT DISTINCT tu.id, tu.model, tu.unit_type, tu.capacity
FROM tractor_unit tu
         LEFT JOIN truck t ON tu.id = t.tractor_unit_id
         LEFT JOIN appointments a ON t.id = a.truck_id
WHERE tu.user_id = $1
  AND (a.situation != 'em andamento' OR a.situation IS NULL);


-- name: GetDashboardTrailerEnterprise :many
SELECT DISTINCT tu.id, 'Carroceiria' as model, tu.body_type, tu.load_capacity
FROM trailer tu
         LEFT JOIN truck t ON tu.id = t.trailer_id
         LEFT JOIN appointments a ON t.id = a.truck_id
WHERE tu.user_id = $1
  AND (a.situation != 'em andamento' OR a.situation IS NULL);


-- name: GetOffersForDashboard :one
SELECT
    cr.advertisement_user_id AS recipient_user_id,
    COALESCE(
            SUM(
                    CASE
                        WHEN EXTRACT(MONTH FROM cm.created_at) = EXTRACT(MONTH FROM CURRENT_DATE)
                            AND EXTRACT(YEAR FROM cm.created_at) = EXTRACT(YEAR FROM CURRENT_DATE)
                            THEN 1
                        ELSE 0
                        END
            ), 0::bigint
    )::bigint AS total_offers_mes_atual,
    COALESCE(
            SUM(
                    CASE
                        WHEN EXTRACT(MONTH FROM cm.created_at) = CASE
                                                                     WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 THEN 12
                                                                     ELSE EXTRACT(MONTH FROM CURRENT_DATE) - 1
                            END
                            AND EXTRACT(YEAR FROM cm.created_at) = CASE
                                                                       WHEN EXTRACT(MONTH FROM CURRENT_DATE) = 1 THEN EXTRACT(YEAR FROM CURRENT_DATE) - 1
                                                                       ELSE EXTRACT(YEAR FROM CURRENT_DATE)
                                END
                            THEN 1
                        ELSE 0
                        END
            ), 0::bigint
    )::bigint AS total_offers_mes_anterior
FROM chat_messages cm
         INNER JOIN chat_rooms cr ON cm.room_id = cr.id
WHERE cm.type_message = 'offer'
  AND cr.advertisement_user_id = $1
GROUP BY cr.advertisement_user_id;
