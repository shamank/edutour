.PHONY:
.SILENT:


up:
	docker build -t edutour/api-gateway ./apigateway
	docker build -t edutour/auth-service ./auth-service
	docker-compose up -d

down:
	docker-compose down