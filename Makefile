.PHONY: up down test run

up:
	docker compose up --build

down:
	docker compose down

test:
	go test ./...

run:
	go run ./cmd/karat-server
