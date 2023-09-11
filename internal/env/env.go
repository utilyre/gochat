package env

import "os"

type Env struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string

	BEPort   string
	BESecret []byte
}

func New() Env {
	return Env{
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),

		BEPort:   os.Getenv("BE_PORT"),
		BESecret: []byte(os.Getenv("BE_SECRET")),
	}
}
