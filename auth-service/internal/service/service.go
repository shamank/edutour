package service

import (
	"context"
	"github.com/patrickmn/go-cache"
	"github.com/shamank/edutour-backend/auth-service/internal/domain"
	"github.com/shamank/edutour-backend/auth-service/internal/repository"
	"github.com/shamank/edutour-backend/auth-service/pkg/auth"
	"github.com/shamank/edutour-backend/auth-service/pkg/email"
	"github.com/shamank/edutour-backend/auth-service/pkg/hash"
	"log/slog"
	"time"
)

type UserSignUpInput struct {
	UserName string
	Email    string
	Phone    string
	Password string
}

type UserSignInInput struct {
	Login    string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpireIn     time.Duration
}

type UserProfile struct {
	UserName   string
	FirstName  string
	LastName   string
	MiddleName string
	Avatar     string
	Role       string
}

type Authorization interface {
	SignUp(ctx context.Context, input UserSignUpInput) error
	SignIn(ctx context.Context, input UserSignInInput) (Tokens, error)
	ConfirmUser(ctx context.Context, confirmToken string) error

	ResetPassword(ctx context.Context, email string) error
	ConfirmResetPassword(ctx context.Context, token string, password string) error

	RefreshToken(ctx context.Context, refreshToken string) (Tokens, error)
	Verify(ctx context.Context, userID int, hash string) error

	setRefreshToken(ctx context.Context, userID int, userName string, userRole string) (Tokens, error)
	GetFullUserInfo(ctx context.Context, userID int) (domain.User, error)
}

type UserProfileInput struct {
	FirstName  string
	LastName   string
	MiddleName string
	Avatar     string
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Users
type Users interface {
	GetUserProfile(ctx context.Context, userName string) (UserProfile, error)
	UpdateUserProfile(ctx context.Context, userName string, user UserProfileInput) error
	ChangeUserPassword(ctx context.Context, userID int, oldPassword, newPassword string) error
}

type Services struct {
	repos         *repository.Repository
	logger        *slog.Logger
	Authorization Authorization
	Users         Users
}

type Dependencies struct {
	Cache        *cache.Cache
	Hasher       hash.PasswordHasher
	TokenManager auth.TokenManager
	EmailManager *email.EmailManager
}

func NewServices(repos *repository.Repository, logger *slog.Logger, dependencies Dependencies) *Services {

	return &Services{
		repos:         repos,
		logger:        logger,
		Authorization: NewAuthService(repos.Authorization, logger, dependencies.Hasher, dependencies.TokenManager, dependencies.EmailManager),
		Users:         NewUserService(repos.Users, logger, dependencies.Hasher),
	}
}
