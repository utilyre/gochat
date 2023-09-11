package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/caarlos0/httperr"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/utilyre/gochat/internal/env"
	"github.com/utilyre/gochat/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8"`
}

type usersHandler struct {
	env      env.Env
	logger   *slog.Logger
	validate *validator.Validate
	storage  storage.UsersStorage
}

func Users(
	r *mux.Router,
	env env.Env,
	logger *slog.Logger,
	validate *validator.Validate,
	storage storage.UsersStorage,
) {
	s := r.PathPrefix("/api/users").Subrouter()
	h := usersHandler{
		env:      env,
		logger:   logger,
		validate: validate,
		storage:  storage,
	}

	s.Handle("/signup", httperr.NewF(h.signup)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	s.Handle("/login", httperr.NewF(h.login)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")
}

func (h usersHandler) signup(w http.ResponseWriter, r *http.Request) error {
	user := new(User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		return httperr.Errorf(http.StatusBadRequest, "%v", err)
	}
	if err := h.validate.Struct(user); err != nil {
		return httperr.Errorf(http.StatusBadRequest, "%v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	dbUser := &storage.User{
		Email:    user.Email,
		Password: hash,
	}

	if err := h.storage.Create(dbUser); err != nil {
		if errors.Is(err, storage.ErrDuplicateKey) {
			return httperr.Errorf(http.StatusConflict, "%v", ErrUserExists)
		}

		return err
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Created"))
	return nil
}

func (h usersHandler) login(w http.ResponseWriter, r *http.Request) error {
	user := new(User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		return httperr.Errorf(http.StatusBadRequest, "%v", err)
	}
	if err := h.validate.Struct(user); err != nil {
		return httperr.Errorf(http.StatusBadRequest, "%v", err)
	}

	dbUser := &storage.User{Email: user.Email}
	if err := h.storage.ReadByEmail(dbUser); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httperr.Errorf(http.StatusNotFound, "%v", ErrUserNotFound)
		}

		return err
	}

	if err := bcrypt.CompareHashAndPassword(dbUser.Password, []byte(user.Password)); err != nil {
		return httperr.Errorf(http.StatusNotFound, "%v", ErrUserNotFound)
	}

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": dbUser.Email,
			"exp":   time.Now().Add(72 * time.Hour).Unix(),
		},
	).SignedString(h.env.BESecret)
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]any{"token": token})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
	return nil
}
