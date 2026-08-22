-- +migrate Up
ALTER TABLE products
    ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_products_user_id
    ON products(user_id)
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_products_user_id;
ALTER TABLE products DROP COLUMN IF EXISTS user_id;
