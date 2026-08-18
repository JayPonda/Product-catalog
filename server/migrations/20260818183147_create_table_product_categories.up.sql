CREATE TABLE product_categories {
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id),
    category_id UUID NOT NULL REFERENCES categories(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ

}

CREATE INDEX 
    idx_product_category_id
ON 
    product_categories (product_id)
WHERE 
    deleted_at 
IS NULL;
