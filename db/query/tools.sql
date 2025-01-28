-- name: CreateTolls :exec
INSERT INTO public.tolls
(concessionaria, praca_de_pedagio, ano_do_pnv_snv, rodovia, uf, km_m, municipio, tipo_pista, sentido, situacao, data_da_inativacao, latitude, longitude)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: GetTollsByLonAndLat :many
SELECT *
FROM public.tolls;
