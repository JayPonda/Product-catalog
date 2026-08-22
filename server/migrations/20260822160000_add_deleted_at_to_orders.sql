-- +migrate Up
ALTER TABLE orders
    ADD COLUMN deleted_at TIMESTAMPTZ;

-- +migrate Down
ALTER TABLE orders
    DROP COLUMN IF EXISTS deleted_at;
