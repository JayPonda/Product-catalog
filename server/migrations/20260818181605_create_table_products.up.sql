CREATE TABLE products {
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(150) NOT NULL;
    price numeric(12,2) NOT NULL;
    stock_quantity INT NOT NULL;
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
}

CREATE INDEX 
    idx_products_name
ON 
    products(name)
WHERE 
    deleted_at 
IS NULL;
