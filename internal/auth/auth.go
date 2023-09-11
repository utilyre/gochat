package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims

	Email string `json:"email"`
}

func Generate(secret []byte, email string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			},
			Email: email,
		},
	)

	return token.SignedString(secret)
}

func Verify(secret []byte, token string) (*Claims, error) {
	claims := new(Claims)

	if _, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) { return secret, nil },
	); err != nil {
		return nil, err
	}

	return claims, nil
}
