SELECT
    id, name, identifier, city, address, created_at, updated_at
FROM
    companies
WHERE
    id = $1;