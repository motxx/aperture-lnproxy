CREATE TABLE IF NOT EXISTS secrets (
    id INTEGER PRIMARY KEY,
    macaroon_id_hash BLOB UNIQUE NOT NULL,
    payment_hash BLOB NOT NULL,
    secret BLOB UNIQUE NOT NULL,
    settled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS secrets_hash_idx ON secrets (macaroon_id_hash);
CREATE INDEX IF NOT EXISTS secrets_hash_idx ON secrets (payment_hash);
