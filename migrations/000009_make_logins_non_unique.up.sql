BEGIN;
    ALTER TABLE "user" DROP CONSTRAINT IF EXISTS user_login_key;

    CREATE UNIQUE INDEX user_login_unique ON "user" (login) WHERE deleted_at IS NULL;
COMMIT;

