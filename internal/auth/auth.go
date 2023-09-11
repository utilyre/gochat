package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/utilyre/gochat/internal/env"
)

type Claims struct {
	jwt.RegisteredClaims

	Email string `json:"email"`
}

type Auth struct {
	env env.Env
}

func New(env env.Env) Auth {
	return Auth{env: env}
}

func (a Auth) Generate(email string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			},
			Email: email,
		},
	)

	return token.SignedString(a.env.BESecret)
}

func (a Auth) Verify(token string) (*Claims, error) {
	claims := new(Claims)

	if _, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) { return a.env.BESecret, nil },
	); err != nil {
		return nil, err
	}

	return claims, nil
}
