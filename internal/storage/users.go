package storage

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

	if err := s.db.Get(user, query, user.Email, user.Password); err != nil {
		pqErr := new(pq.Error)
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrDuplicateKey
		}

		return err
	}

	return nil
}

func (s UsersStorage) ReadByEmail(user *User) error {
	query := `
	SELECT "id", "created_at", "updated_at", "password"
	FROM "users"
	WHERE "email" = $1;
	`

	return s.db.Get(user, query, user.Email)
}
