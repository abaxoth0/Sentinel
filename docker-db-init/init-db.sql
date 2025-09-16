SELECT 'CREATE DATABASE sentinel' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'sentinel')\gexec
CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD '1234'; -- TODO change in prod

-- This is just all 'up' migrations
BEGIN;
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

    CREATE INDEX IF NOT EXISTS user_pagination_idx on "user" (created_at DESC, id DESC);

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

    CREATE TABLE IF NOT EXISTS location (
        id          UUID PRIMARY KEY,
        ip          INET NOT NULL,
        session_id  UUID REFERENCES "user_session"(id) ON DELETE SET NULL,
        country     VARCHAR(2) NOT NULL, -- ISO 3166-1 alpha-2
        region      VARCHAR(3),          -- ISO 3166-2 region code
        city        VARCHAR(100),
        latitude    REAL,
        longitude   REAL,
        isp         VARCHAR(100),
        deleted_at  TIMESTAMP,
        created_at  TIMESTAMP DEFAULT NOW() NOT NULL
    );

    ALTER TABLE "user_session" DROP COLUMN IF EXISTS location;

    ALTER TABLE IF EXISTS "user_session" RENAME COLUMN revoked TO revoked_at;

    ALTER TABLE "user_session" ADD COLUMN revoked_temp TIMESTAMP;

    UPDATE "user_session" SET revoked_temp = CURRENT_TIMESTAMP WHERE revoked_at = true;

    ALTER TABLE "user_session" DROP COLUMN revoked_at;

    ALTER TABLE "user_session" RENAME COLUMN revoked_temp TO revoked_at;

    ALTER TABLE IF EXISTS "audit_user" ADD COLUMN reason TEXT;

    CREATE TABLE IF NOT EXISTS "audit_location" (
        id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        changed_location_id uuid REFERENCES "location"(id) ON DELETE CASCADE,
        operation           CHAR(1) NOT NULL,
        ip                  INET NOT NULL,
        session_id          UUID REFERENCES "user_session"(id) ON DELETE SET NULL,
        country             VARCHAR(2) NOT NULL, -- ISO 3166-1 alpha-2
        region              VARCHAR(3),          -- ISO 3166-2 region code
        city                VARCHAR(100),
        latitude            REAL,
        longitude           REAL,
        isp                 VARCHAR(100),
        deleted_at          TIMESTAMP,
        created_at          TIMESTAMP NOT NULL,
        changed_at          TIMESTAMP NOT NULL DEFAULT NOW()
    );

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

    ALTER TABLE "user" DROP CONSTRAINT IF EXISTS user_login_key;

    CREATE UNIQUE INDEX user_login_unique ON "user" (login) WHERE deleted_at IS NULL;
COMMIT;

