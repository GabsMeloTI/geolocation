-- name: ConsultUniqueCEP :one
SELECT * FROM unique_ceps WHERE cep = $1::varchar;