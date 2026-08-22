-- +migrate Up
CREATE INDEX 
    idx_product_category_id
ON 
    product_categories (product_id)
WHERE 
    deleted_at 
IS NULL;

-- +migrate Down
DROP INDEX 
    IF EXISTS 
        idx_product_category_id;
