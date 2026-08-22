-- +migrate Up
CREATE INDEX
    idx_orders_customer_id
ON
    orders (customer_id);

CREATE INDEX
    idx_order_products_order_id
ON
    order_products (order_id)
WHERE
    deleted_at
IS NULL;

CREATE INDEX
    idx_order_products_product_id
ON
    order_products (product_id)
WHERE
    deleted_at
IS NULL;

CREATE UNIQUE INDEX
    uq_order_products_order_product_active
ON
    order_products (order_id, product_id)
WHERE
    deleted_at
IS NULL;

-- +migrate Down
DROP INDEX
    IF EXISTS
        uq_order_products_order_product_active;

DROP INDEX
    IF EXISTS
        idx_order_products_product_id;

DROP INDEX
    IF EXISTS
        idx_order_products_order_id;

DROP INDEX
    IF EXISTS
        idx_orders_customer_id;
