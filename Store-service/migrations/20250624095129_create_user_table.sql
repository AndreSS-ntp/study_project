-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_group AS ENUM (
    'admin',
    'customer'
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    last_name VARCHAR(50) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    middle_name VARCHAR(50),
    user_group user_group NOT NULL
);

COMMENT ON COLUMN users.user_group IS 'Группа для разграничения прав доступа: admin - администратор, customer - обычный пользователь';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_group;
-- +goose StatementEnd
