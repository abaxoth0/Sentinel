BEGIN;
    ALTER TABLE "user_session" ADD COLUMN revoked_temp BOOLEAN;

    UPDATE "user_session" SET revoked_temp = (revoked_at IS NOT NULL);

    ALTER TABLE "user_session" DROP COLUMN revoked_at;

    ALTER TABLE "user_session" RENAME COLUMN revoked_temp TO revoked;

    ALTER TABLE "user_session" ALTER COLUMN revoked SET DEFAULT false;
COMMIT;
