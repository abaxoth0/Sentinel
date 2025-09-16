BEGIN;
    DROP INDEX IF EXISTS user_login_unique;

    -- If there are any duplicate logins, the this operation will fail.
    -- You'll need to handle these conflicts manually before running the down migration.
    ALTER TABLE "user" ADD CONSTRAINT user_login_key UNIQUE (login);
COMMIT;

