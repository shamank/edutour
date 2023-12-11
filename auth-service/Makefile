include .env
export $(shell sed 's/=.*//' .env)
.PHONY:
.SILENT:

build:
	go mod download && go build -o ./.bin/app ./cmd/main.go

run: build
	./.bin/app

start:
	go run ./cmd/main.go

docker-build:
	docker build -t edutour/auth-service .

docker-run:
	docker run edutour/auth-service

compose:
	docker-compose up -d

compose-down:
	docker-compose down

rl: compose-down docker-build compose

make swag:
	swag init -g internal/app/app.go

migrate-up:
	migrate -path ./migrations -database 'postgres://pguser:${DB_PASSWORD}@localhost:5431/devdb?sslmode=disable' up

migrate-down:
	migrate -path ./migrations -database 'postgres://pguser:${DB_PASSWORD}@localhost:5431/devdb?sslmode=disable' down

