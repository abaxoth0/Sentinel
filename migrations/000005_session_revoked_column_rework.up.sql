BEGIN;
    ALTER TABLE IF EXISTS "user_session" RENAME COLUMN revoked TO revoked_at;

    ALTER TABLE "user_session" ADD COLUMN revoked_temp TIMESTAMP;

    UPDATE "user_session" SET revoked_temp = CURRENT_TIMESTAMP WHERE revoked_at = true;

    ALTER TABLE "user_session" DROP COLUMN revoked_at;

    ALTER TABLE "user_session" RENAME COLUMN revoked_temp TO revoked_at;
COMMIT;
