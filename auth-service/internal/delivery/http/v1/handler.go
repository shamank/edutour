package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shamank/edutour-backend/auth-service/internal/service"
	"github.com/shamank/edutour-backend/auth-service/pkg/auth"
	"log/slog"
)

type Handler struct {
	services     *service.Services
	logger       *slog.Logger
	tokenManager auth.TokenManager
	validator    *validator.Validate
}

func NewHandler(services *service.Services, logger *slog.Logger, tokenManager auth.TokenManager) *Handler {
	validate := validator.New()
	return &Handler{
		services:     services,
		logger:       logger,
		tokenManager: tokenManager,
		validator:    validate,
	}
}

func (h *Handler) InitAPI(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initAuthRouter(v1)
		h.initUsersRouter(v1)
	}
}
