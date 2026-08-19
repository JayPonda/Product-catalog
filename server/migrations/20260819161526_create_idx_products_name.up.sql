
CREATE INDEX 
    idx_products_name
ON 
    products(name)
WHERE 
    deleted_at 
IS NULL;