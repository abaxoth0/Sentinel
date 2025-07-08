BEGIN;
    DROP TABLE IF EXISTS "user_session";

    ALTER TABLE "user" DROP COLUMN IF EXISTS version;

    ALTER TABLE "audit_user" DROP COLUMN IF EXISTS version;
COMMIT;
