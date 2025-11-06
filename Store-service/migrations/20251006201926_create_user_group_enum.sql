-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_group AS ENUM ('admin','customer');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS user_group;
-- +goose StatementEnd
