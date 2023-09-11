include .env
export

dev:
	gow -s -e go,mod,html,css,js run cmd/main.go
