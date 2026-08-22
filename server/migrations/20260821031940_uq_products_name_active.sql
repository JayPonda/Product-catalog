-- +migrate Up
CREATE UNIQUE INDEX uq_products_name_active
ON products(name)
WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX IF EXISTS uq_products_name_active;
