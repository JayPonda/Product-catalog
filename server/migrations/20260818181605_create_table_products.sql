-- +migrate Up
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(150) NOT NULL,
    price numeric(12,2) NOT NULL,
    stock_quantity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- +migrate Down
DROP TABLE 
    IF EXISTS
        products;
