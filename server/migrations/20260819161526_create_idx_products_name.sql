-- +migrate Up
CREATE INDEX 
    idx_products_name
ON 
    products(name)
WHERE 
    deleted_at 
IS NULL;

-- +migrate Down
DROP INDEX 
    IF EXISTS  
        idx_products_name;
