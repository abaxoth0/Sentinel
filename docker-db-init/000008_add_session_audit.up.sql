BEGIN;
    CREATE TABLE IF NOT EXISTS "audit_user_session" (
        id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        changed_session_id  UUID REFERENCES "user_session"(id) ON DELETE CASCADE,
        changed_by_user_id  UUID REFERENCES "user"(id) ON DELETE SET NULL,
        operation           CHAR(1) NOT NULL,
        user_id             UUID REFERENCES "user"(id) ON DELETE SET NULL,
        user_agent          TEXT NOT NULL,
        ip_address          INET,
        device_id           TEXT,
        device_type         TEXT NOT NULL,
        os                  TEXT NOT NULL,
        os_version          TEXT,
        browser             TEXT NOT NULL,
        browser_version     TEXT,
        created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
        last_used_at        TIMESTAMP,
        expires_at          TIMESTAMP NOT NULL,
        revoked_at          TIMESTAMP,
        changed_at          TIMESTAMP NOT NULL DEFAULT NOW(),
        reason              TEXT
    );
COMMIT;

