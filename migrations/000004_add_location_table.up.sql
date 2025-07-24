BEGIN;
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
COMMIT;
