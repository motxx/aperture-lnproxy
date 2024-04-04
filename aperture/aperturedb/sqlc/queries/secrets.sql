-- name: InsertSecret :one
INSERT INTO secrets (
    macaroon_id_hash, payment_hash, secret, created_at
) VALUES (
    $1, $2, $3, $4
) RETURNING id;

-- name: GetSecretByIdHash :one
SELECT secret
FROM secrets
WHERE macaroon_id_hash = $1;


-- name: GetSettledAtByPaymentHash :one
SELECT settled_at
FROM secrets
WHERE payment_hash = $1;

-- name: SetSettledAtByPaymentHash :exec
UPDATE secrets
SET settled_at = $2
WHERE payment_hash = $1;

-- name: DeleteSecretByIdHash :execrows
DELETE FROM secrets
WHERE macaroon_id_hash = $1;
