package repository

import (
	"context"
	"database/sql"
	"github.com/shamank/edutour-backend/auth-service/internal/domain"
	"github.com/shamank/edutour-backend/auth-service/internal/repository/postgres"
	"log/slog"
)

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

type Authorization interface {
	Create(ctx context.Context, user domain.User, confirmToken string, expireAt int64) error
	GetByCredentials(ctx context.Context, email string, passwordHash string) (domain.User, error)
	GetByUsername(ctx context.Context, username string, passwordHash string) (domain.User, error)

	SetTokenResetPassword(ctx context.Context, email string, token string, expireAt int64) error
	ConfirmResetPassword(ctx context.Context, token string, passwordHash string) error

	ConfirmUser(ctx context.Context, confirmToken string) error

	GetByRefreshToken(ctx context.Context, refreshToken string) (domain.User, error)

	SetRefreshToken(ctx context.Context, userID int, refreshInput domain.RefreshToken) error
	Verify(ctx context.Context, userID int) error
	GetFullUserInfo(ctx context.Context, userID int) (domain.User, error)
}

type Users interface {
	GetUserProfile(ctx context.Context, userName string) (domain.User, error)
	UpdateUserProfile(ctx context.Context, user domain.User) error
	ChangeUserPassword(ctx context.Context, userID int, oldPasswordHash, newPasswordHash string) error
}

type Migrator interface {
	Up(migrationPath string) error
	Down(migrationPath string) error
}

type Repository struct {
	db            *sql.DB
	logger        *slog.Logger
	Authorization Authorization
	Users         Users
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{
		db:            db,
		logger:        logger,
		Authorization: postgres.NewAuthRepo(db, logger),
		Users:         postgres.NewUserRepo(db, logger),
	}
}
