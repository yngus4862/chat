.PHONY: build run migrate-up migrate-down

build:
	go build -o bin/chatd ./cmd/chatd

run:
	go run ./cmd/chatd

migrate-up:
	migrate -path migrations -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" down 1