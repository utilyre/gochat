-- +goose Up
-- +goose StatementBegin
CREATE TABLE "users_rooms" (
    "user_id" serial REFERENCES "users"("id"),
    "room_id" serial REFERENCES "rooms"("id"),
    PRIMARY KEY ("user_id", "room_id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "users_rooms";
-- +goose StatementEnd
