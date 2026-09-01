-- +migrate Up
CREATE INDEX IF NOT EXISTS
    idx_product_categories_category_id
ON
    product_categories (category_id)
WHERE
    deleted_at
IS NULL;

-- +migrate Down
DROP INDEX
    IF EXISTS
        idx_product_categories_category_id;
