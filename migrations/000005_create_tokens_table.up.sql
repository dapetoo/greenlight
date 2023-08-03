CREATE TABLE IF NOT EXISTS tokens(
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES ON users ON DELETE CASCADE,
    expiry TIMESTAMP(0)  WITH time ZONE NOT NUL,
    scope text NOT NULL
);