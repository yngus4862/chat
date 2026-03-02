.PHONY: run build smoke migrate-up migrate-down chatctl test

build:
	go build -o ./tmp/chatd ./cmd/chatd

run:
	go run ./cmd/chatd

smoke:
	go run ./cmd/smoketest -api http://127.0.0.1:8080 -ws ws://127.0.0.1:8081/ws

chatctl:
	@echo "usage: make chatctl CMD=status|stop|restart TOKEN=... ADDR=http://127.0.0.1:9099"
	go run ./cmd/chatctl -addr $(ADDR) -token $(TOKEN) $(CMD)

migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

test:
	go test ./... -v