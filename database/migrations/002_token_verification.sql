-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN verification_token TEXT;
CREATE INDEX idx_users_verification_token ON users(verification_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_users_verification_token;
ALTER TABLE users DROP COLUMN verification_token;
-- +goose StatementEnd
