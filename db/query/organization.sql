-- name: GetOrganizationByTenant :one
SELECT postcode
FROM "Organizations"
WHERE access_id=$1 AND
      tenant_id=$2 AND
      CNPJ=$3;