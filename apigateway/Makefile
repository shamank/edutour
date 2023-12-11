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

prod: run
	--prod

docker-build:
	docker build -t edutour/api-gateway .

docker-run:
	docker run edutour/api-gateway

compose:
	docker-compose up -d

compose-down:
	docker-compose down

rl: compose-down docker-build compose

