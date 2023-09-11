package storage

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrDuplicateKey = errors.New("duplicate key value violates unique constraint")
)

type User struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Email    string `db:"email"`
	Password []byte `db:"password"`
}

type UsersStorage struct {
	db *sqlx.DB
}

func NewUsersStorage(db *sqlx.DB) UsersStorage {
	return UsersStorage{db: db}
}

func (s UsersStorage) Create(user *User) error {
	query := `
	INSERT
	INTO "users"
	("email", "password")
	VALUES
		($1, $2)
	RETURNING "id", "created_at", "updated_at";
	`

	return s.db.Get(user, query, user.Email, user.Password)
}

func (s UsersStorage) ReadByEmail(user *User) error {
	query := `
	SELECT "id", "created_at", "updated_at", "password"
	FROM "users"
	WHERE "email" = $1;
	`

	return s.db.Get(user, query, user.Email)
}
