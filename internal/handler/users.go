package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/utilyre/gochat/internal/auth"
	"github.com/utilyre/gochat/internal/env"
	"github.com/utilyre/gochat/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64  `json:"id" validate:"isdefault"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password,omitempty" validate:"required,min=8"`
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

	s.HandleFunc("/signup", h.signup).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	s.HandleFunc("/login", h.login).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")
}

func (h usersHandler) signup(w http.ResponseWriter, r *http.Request) {
	user := new(User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Warn("failed to generate hash from password", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dbUser := &storage.User{
		Email:    user.Email,
		Password: hash,
	}

	if err := h.storage.Create(dbUser); err != nil {
		switch {
		case errors.Is(err, storage.ErrDuplicateKey):
			http.Error(w, "user already exists", http.StatusConflict)
		default:
			h.logger.Warn("failed to create user in users table", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	user.ID = dbUser.ID
	user.Password = ""

	body, err := json.Marshal(user)
	if err != nil {
		h.logger.Warn("failed to marshal response body", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(body); err != nil {
		h.logger.Warn("failed to write body to response", "error", err)
	}
}

func (h usersHandler) login(w http.ResponseWriter, r *http.Request) {
	user := new(User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbUser := &storage.User{Email: user.Email}
	if err := h.storage.ReadByEmail(dbUser); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.Error(w, "user not found", http.StatusNotFound)
		default:
			h.logger.Warn("failed to read user by email", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	if err := bcrypt.CompareHashAndPassword(dbUser.Password, []byte(user.Password)); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	token, err := auth.Generate(h.env.BESecret, dbUser.Email)
	if err != nil {
		h.logger.Warn("failed to generate JWT", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(map[string]any{"token": token})
	if err != nil {
		h.logger.Warn("failed to marshal response body", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(body); err != nil {
		h.logger.Warn("failed to write body to response", "error", err)
	}
}
