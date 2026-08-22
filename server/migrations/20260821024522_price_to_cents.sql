-- +migrate Up
ALTER TABLE products
    ALTER COLUMN price TYPE BIGINT USING ROUND(price * 100)::BIGINT;

-- +migrate Down
ALTER TABLE products
    ALTER COLUMN price TYPE NUMERIC(12,2) USING (price::NUMERIC(12,2) / 100);
