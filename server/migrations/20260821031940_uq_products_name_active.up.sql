CREATE UNIQUE INDEX uq_products_name_active
ON products(name)
WHERE deleted_at IS NULL;
