FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./.bin/app ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/.bin/app .bin/app
COPY --from=builder /app/configs configs/
COPY --from=builder /app/.env .
COPY --from=builder /app/migrations migrations/



ENV DOCKERIZE_VERSION v0.7.0

RUN apk update --no-cache \
    && apk add --no-cache wget openssl \
    && wget -O - https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz | tar xzf - -C /usr/local/bin \
    && apk del wget

EXPOSE 8000

CMD ["dockerize", "-wait", "tcp://postgres:5432", "-timeout", "15s", ".bin/app", "--prod"]

