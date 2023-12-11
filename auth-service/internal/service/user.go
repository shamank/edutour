package service

import (
	"context"
	"github.com/shamank/edutour-backend/auth-service/internal/domain"
	"github.com/shamank/edutour-backend/auth-service/internal/repository"
	"github.com/shamank/edutour-backend/auth-service/pkg/hash"
	"log/slog"
)

type UserService struct {
	repo   repository.Users
	logger *slog.Logger
	hasher hash.PasswordHasher
}

func NewUserService(repo repository.Users, logger *slog.Logger, hasher hash.PasswordHasher) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
		hasher: hasher,
	}
}

func (s *UserService) GetUserProfile(ctx context.Context, userName string) (UserProfile, error) {

	res, err := s.repo.GetUserProfile(ctx, userName)
	if err != nil {
		return UserProfile{}, err
	}

	return UserProfile{
		FirstName:  res.FirstName,
		LastName:   res.LastName,
		MiddleName: res.MiddleName,
		Avatar:     res.Avatar,
		Role:       res.Role.Name,
	}, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userName string, user UserProfileInput) error {
	err := s.repo.UpdateUserProfile(ctx, domain.User{
		Username:   userName,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Avatar:     user.Avatar,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) ChangeUserPassword(ctx context.Context, userID int, oldPassword, newPassword string) error {

	// TODO: добавить логгер & обработку ошибок
	oldPasswordHash, err := s.hasher.Hash(oldPassword)
	if err != nil {
		return err
	}
	newPasswordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	err = s.repo.ChangeUserPassword(ctx, userID, oldPasswordHash, newPasswordHash)
	if err != nil {
		return err
	}

	return nil
}
