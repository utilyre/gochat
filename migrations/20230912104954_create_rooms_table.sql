-- +goose Up
-- +goose StatementBegin
CREATE TABLE "rooms" (
    "id" serial NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

    "name" varchar(50) NOT NULL UNIQUE,
    PRIMARY KEY ("id")
);

CREATE TRIGGER "refresh_rooms_updated_at"
    BEFORE UPDATE
    ON "rooms"
    FOR EACH ROW
    EXECUTE PROCEDURE refresh_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER "refresh_rooms_updated_at" ON "rooms";
DROP TABLE "rooms";
-- +goose StatementEnd
