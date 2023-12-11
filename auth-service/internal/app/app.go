package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"github.com/shamank/edutour-backend/auth-service/internal/config"
	handler "github.com/shamank/edutour-backend/auth-service/internal/delivery/http"
	"github.com/shamank/edutour-backend/auth-service/internal/repository"
	"github.com/shamank/edutour-backend/auth-service/internal/server"
	"github.com/shamank/edutour-backend/auth-service/internal/service"
	"github.com/shamank/edutour-backend/auth-service/pkg/auth"
	"github.com/shamank/edutour-backend/auth-service/pkg/email"
	"github.com/shamank/edutour-backend/auth-service/pkg/hash"
	"github.com/shamank/edutour-backend/auth-service/pkg/logger/sl"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title EduTour-AuthService API
// @version 1.0
// @description REST API for EduTour-AuthService

// @host 109.172.81.237:8000
// @BasePath /api/v1/

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

const (
	envLocal = "local"
	envProd  = "prod"
)

func Run(configDir string) {

	//cfg, err := config.Init(configDir, "local")
	cfg := config.InitConfig(configDir)

	logger := setupLogger(cfg.Env)

	logger.Info("starting auth-service", slog.String("env", cfg.Env))

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode))
	if err != nil {
		logger.Error("error occurred in connecting to postgres", sl.Err(err))
	}

	if cfg.MigrationPath != "" {
		if err := checkMigrations(db, cfg.MigrationPath); err != nil {
			logger.Error(err.Error())
			return
		}
	}

	repos := repository.NewRepository(db, logger)

	memcache := cache.New(5*time.Minute, 10*time.Minute)

	JWTConfig := cfg.AuthConfig.JWT

	tokenManager, err := auth.NewManager(JWTConfig.SignedKey, JWTConfig.AccessTokenTTL, JWTConfig.RefreshTokenTTL)
	if err != nil {
		logger.Error("error occurred generate tokenManager", sl.Err(err))
	}

	SMTP := email.NewSMTPServer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Password)
	emailManager := email.NewEmailManager(SMTP, cfg.SMTP.User)

	hasher := hash.NewSHA256Hasher(cfg.AuthConfig.PasswordSalt)

	deps := service.Dependencies{
		Cache:        memcache,
		Hasher:       hasher,
		TokenManager: tokenManager,
		EmailManager: emailManager,
	}

	services := service.NewServices(repos, logger, deps)

	handlers := handler.NewHandler(services, logger, tokenManager)

	srv := server.NewServer(cfg, handlers.InitAPI())

	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("error occurred when starting the HTTP-server", sl.Err(err))
			return
		}
		logger.Info("")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("error occurred when stopping the HTTP server", sl.Err(err))
		return
	}

	logger.Info("the server has shut down")
}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
	}
	return log
}

func checkMigrations(db *sql.DB, migrationPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
	return nil
}
