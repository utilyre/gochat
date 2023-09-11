package env

import "os"

type Env struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string

	BEPort string
}

func New() Env {
	return Env{
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),

		BEPort: os.Getenv("BE_PORT"),
	}
}
