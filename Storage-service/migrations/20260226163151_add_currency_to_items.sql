-- +goose Up
-- +goose StatementBegin
ALTER TABLE items ADD COLUMN currency VARCHAR(3);
UPDATE items SET currency = 'RUB' WHERE currency IS NULL;
ALTER TABLE items ALTER COLUMN currency SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE items DROP COLUMN IF EXISTS currency;
-- +goose StatementEnd
