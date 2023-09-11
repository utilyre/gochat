-- +goose Up
-- +goose StatementBegin
CREATE TABLE "users" (
    "id" serial NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

    "email" varchar(255) NOT NULL UNIQUE,
    "password" bytea NOT NULL,
    PRIMARY KEY ("id")
);

CREATE FUNCTION refresh_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER "refresh_users_updated_at"
    BEFORE UPDATE
    ON "users"
    FOR EACH ROW
    EXECUTE PROCEDURE refresh_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER "refresh_users_updated_at" ON "users";
DROP FUNCTION "refresh_updated_at";
DROP TABLE "users";
-- +goose StatementEnd
