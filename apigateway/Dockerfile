FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./.bin/app ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/.bin/app .bin/app
COPY --from=builder /app/configs configs/

EXPOSE 8000

CMD [".bin/app", "--prod"]