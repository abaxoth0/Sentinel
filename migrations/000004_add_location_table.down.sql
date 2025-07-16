BEGIN;
    DROP TABLE IF EXISTS "location";

    ALTER TABLE "user_session" ADD COLUMN IF NOT EXISTS location TEXT;
COMMIT;
