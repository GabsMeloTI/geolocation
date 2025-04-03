CREATE OR REPLACE VIEW view_route_request AS
SELECT rh.id_user, u.name, SUM(rh.number_request) AS total_request
FROM route_hist rh
         INNER JOIN users u ON u.id = rh.id_user
GROUP BY rh.id_user, u.name;


CREATE OR REPLACE VIEW view_route_hist AS
SELECT rh.*, u.name
FROM route_hist rh
         INNER JOIN users u ON u.id = rh.id_user
ORDER BY rh.created_at;
