BEGIN;
    ALTER TABLE audit_user
        ADD CONSTRAINT audit_user_changed_user_id_fkey
            FOREIGN KEY (changed_user_id) REFERENCES "user"(id) ON DELETE CASCADE,
        ADD CONSTRAINT audit_user_changed_by_user_id_fkey
            FOREIGN KEY (changed_by_user_id) REFERENCES "user"(id) ON DELETE CASCADE;

    ALTER TABLE user_session
        ADD CONSTRAINT user_session_changed_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE;

    ALTER TABLE audit_location
        ADD CONSTRAINT audit_location_changed_location_id_fkey
            FOREIGN KEY (location_id) REFERENCES "location"(id) ON DELETE CASCADE;

    ALTER TABLE audit_user_session
        ADD CONSTRAINT audit_user_session_changed_user_session_id_fkey
            FOREIGN KEY (changed_session_id) REFERENCES "user_session"(id) ON DELETE CASCADE,
        ADD CONSTRAINT audit_user_session_changed_changed_by_user_id_fkey
            FOREIGN KEY (changed_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
        ADD CONSTRAINT audit_user_session_changed_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE SET NULL;
COMMIT;

