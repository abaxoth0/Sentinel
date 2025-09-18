BEGIN;
    ALTER TABLE audit_user
        DROP CONSTRAINT IF EXISTS audit_user_changed_user_id_fkey,
        DROP CONSTRAINT IF EXISTS audit_user_changed_by_user_id_fkey;

    ALTER TABLE user_session
        DROP CONSTRAINT IF EXISTS user_session_changed_user_id_fkey;

    ALTER TABLE audit_location
        DROP CONSTRAINT IF EXISTS audit_location_changed_location_id_fkey;

    ALTER TABLE audit_user_session
        DROP CONSTRAINT IF EXISTS audit_user_session_changed_user_session_id_fkey,
        DROP CONSTRAINT IF EXISTS audit_user_session_changed_changed_by_user_id_fkey,
        DROP CONSTRAINT IF EXISTS audit_user_session_changed_user_id_fkey;
COMMIT;

