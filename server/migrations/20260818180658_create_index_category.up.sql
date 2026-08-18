CREATE 
    INDEX 
        idx_categories_lower_name
    ON 
        categories (LOWER(name))
    WHERE 
        deleted_at 
    IS NULL;