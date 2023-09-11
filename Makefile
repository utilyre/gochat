include .env
export

dev:
	gow -s -e go,mod,html,css,js run cmd/main.go

db:
	docker run \
		-p ${DB_PORT}:5432 \
		-v gochat:/var/lib/postgresql/data \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_PASSWORD=${DB_PASS} \
		--name gochat_db \
		postgres:15.3-alpine3.18
