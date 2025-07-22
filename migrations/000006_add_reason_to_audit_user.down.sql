BEGIN;
    ALTER TABLE "audit_user" DROP COLUMN reason;
COMMIT;
