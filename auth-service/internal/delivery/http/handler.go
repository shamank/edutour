package http

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/shamank/edutour-backend/auth-service/internal/delivery/http/v1"
	"github.com/shamank/edutour-backend/auth-service/internal/service"
	"github.com/shamank/edutour-backend/auth-service/pkg/auth"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log/slog"

	_ "github.com/shamank/edutour-backend/auth-service/docs"
)

type Handler struct {
	services     *service.Services
	logger       *slog.Logger
	tokenManager auth.TokenManager
}

func NewHandler(services *service.Services, logger *slog.Logger, tokenManager auth.TokenManager) *Handler {
	return &Handler{
		services:     services,
		logger:       logger,
		tokenManager: tokenManager,
	}
}

func (h *Handler) InitAPI() *gin.Engine {
	router := gin.Default()

	router.Use(CORS)

	handlerV1 := v1.NewHandler(h.services, h.logger, h.tokenManager)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		handlerV1.InitAPI(api)
	}

	return router

}
