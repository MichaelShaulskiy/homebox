-- +goose Up
ALTER TABLE users ADD COLUMN account_role text NOT NULL DEFAULT 'user'
    CHECK (account_role IN ('superadmin', 'admin', 'user'));
UPDATE users SET account_role = 'superadmin' WHERE is_superuser = 1 OR superuser = 1;

CREATE TABLE IF NOT EXISTS user_supervisors (
    user_id uuid NOT NULL,
    subordinate_id uuid NOT NULL,
    PRIMARY KEY (user_id, subordinate_id),
    CONSTRAINT user_supervisors_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT user_supervisors_supervisor FOREIGN KEY (subordinate_id)
        REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS user_supervisors_subordinate_id
    ON user_supervisors(subordinate_id);

-- +goose Down
DROP TABLE IF EXISTS user_supervisors;
ALTER TABLE users DROP COLUMN account_role;
