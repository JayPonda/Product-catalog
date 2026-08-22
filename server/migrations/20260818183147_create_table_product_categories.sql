-- +migrate Up
CREATE TABLE product_categories (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- +migrate Down
DROP TABLE 
    IF EXISTS 
        product_categories;
