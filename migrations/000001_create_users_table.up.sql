CREATE TABLE IF NOT EXISTS "user" (
    id uuid PRIMARY KEY,
    login VARCHAR(72) UNIQUE NOT NULL,
    password CHAR(60) NOT NULL,
    roles VARCHAR(32)[] NOT NULL,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "audit_user" (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    changed_user_id uuid REFERENCES "user"(id) ON DELETE CASCADE,
    changed_by_user_id uuid REFERENCES "user"(id) ON DELETE CASCADE,
    operation CHAR(1) NOT NULL,
    login VARCHAR(72) NOT NULL,
    password CHAR(60) NOT NULL,
    roles VARCHAR(32)[] NOT NULL,
    deleted_at TIMESTAMP,
    changed_at TIMESTAMP NOT NULL
);

