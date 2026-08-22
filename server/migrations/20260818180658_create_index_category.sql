-- +migrate Up
CREATE 
    INDEX 
        idx_categories_lower_name
    ON 
        categories (LOWER(name))
    WHERE 
        deleted_at 
    IS NULL;

-- +migrate Down
DROP INDEX 
    IF EXISTS 
        idx_categories_lower_name;
