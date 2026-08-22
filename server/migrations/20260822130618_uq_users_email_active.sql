-- +migrate Up
CREATE UNIQUE INDEX uq_users_email_active
ON users(email)
WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX IF EXISTS uq_users_email_active;
