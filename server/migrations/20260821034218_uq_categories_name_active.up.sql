CREATE UNIQUE INDEX uq_categories_name_active
ON categories(name)
WHERE deleted_at IS NULL;
