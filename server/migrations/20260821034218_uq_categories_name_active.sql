-- +migrate Up
CREATE UNIQUE INDEX uq_categories_name_active
ON categories(name)
WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX IF EXISTS uq_categories_name_active;
