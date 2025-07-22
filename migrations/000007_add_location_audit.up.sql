BEGIN;
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
        deleted_at          TIMESTAMPTZ,
        created_at          TIMESTAMPTZ NOT NULL
    );
COMMIT;
