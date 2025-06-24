BEGIN;
    CREATE INDEX IF NOT EXISTS user_pagination_idx on "user" (created_at DESC, id DESC);
COMMIT;
