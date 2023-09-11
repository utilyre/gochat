package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/caarlos0/httperr"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/utilyre/gochat/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8"`
}

type usersHandler struct {
	validate *validator.Validate
	storage  storage.UsersStorage
}

func Users(
	r *mux.Router,
	logger *slog.Logger,
	validate *validator.Validate,
	storage storage.UsersStorage,
) {
	s := r.PathPrefix("/api/users").Subrouter()
	h := usersHandler{
		validate: validate,
		storage:  storage,
	}

	s.Handle("/signup", httperr.NewF(h.signup)).
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
		// TODO: dup
		return err
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Created"))
	return nil
}
