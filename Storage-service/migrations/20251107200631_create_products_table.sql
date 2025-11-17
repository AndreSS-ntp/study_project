-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS items (
    sku NUMERIC(20, 0) PRIMARY KEY,
    name TEXT NOT NULL,
    price NUMERIC(19, 4) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity >= 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
-- +goose StatementEnd
