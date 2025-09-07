BEGIN;
    CREATE TABLE IF NOT EXISTS user_session (
        id                  UUID PRIMARY KEY,
        user_id             UUID REFERENCES "user"(id) ON DELETE CASCADE,
        user_agent          TEXT NOT NULL,
        ip_address          INET,
        device_id           TEXT,
        device_type         TEXT NOT NULL,
        os                  TEXT NOT NULL,
        os_version          TEXT,
        browser             TEXT NOT NULL,
        browser_version     TEXT,
        location            TEXT,
        created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
        last_used_at        TIMESTAMP,
        expires_at          TIMESTAMP NOT NULL,
        revoked             BOOL NOT NULL DEFAULT FALSE
    );

    ALTER TABLE "user" ADD COLUMN IF NOT EXISTS version INT DEFAULT 1;

    ALTER TABLE "audit_user" ADD COLUMN IF NOT EXISTS version INT DEFAULT 1;
COMMIT;
