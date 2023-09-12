package storage

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Room struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Name string `db:"name"`
}

type RoomsStorage struct {
	db *sqlx.DB
}

func NewRoomsStorage(db *sqlx.DB) RoomsStorage {
	return RoomsStorage{db: db}
}

func (s RoomsStorage) Create(room *Room) error {
	query := `
	INSERT
	INTO "rooms"
	("name")
	VALUES
		($1)
	RETURNING "id", "created_at", "updated_at";
	`

	if err := s.db.Get(room, query, room.Name); err != nil {
		pqErr := new(pq.Error)
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrDuplicateKey
		}

		return err
	}

	return nil
}

func (s RoomsStorage) ReadAll(rooms *[]Room) error {
	query := `
	SELECT "id", "created_at", "updated_at", "name"
	FROM "rooms";
	`

	return s.db.Select(rooms, query)
}

func (s RoomsStorage) ReadByID(room *Room) error {
	query := `
	SELECT "created_at", "updated_at", "name"
	FROM "rooms"
	WHERE "id" = $1;
	`

	return s.db.Get(room, query, room.ID)
}
