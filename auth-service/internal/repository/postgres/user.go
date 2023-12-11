package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/shamank/edutour-backend/auth-service/internal/domain"
	"github.com/shamank/edutour-backend/auth-service/pkg/logger/sl"
	"log/slog"
	"strings"
)

type UserRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepo(db *sql.DB, logger *slog.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepo) GetUserProfile(ctx context.Context, userName string) (domain.User, error) {
	const op = "Repository.Postgres.UserRepo.GetUserProfile"
	logger := r.logger.With(slog.String("op", op))

	var user domain.User

	query := `SELECT u.username, COALESCE(u.first_name, '') as first_name,
       COALESCE(u.last_name, '') as last_name, COALESCE(u.middle_name, '') as middle_name,
       COALESCE(u.avatar, '') as avatar, r.id, r.name
				FROM users u
				INNER JOIN role_types r on r.id = u.role_id
				WHERE u.username = $1`

	err := r.db.QueryRow(query, userName).Scan(
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Avatar,
		&user.Role.ID,
		&user.Role.Name)

	if err != nil {
		logger.Error("error occurred when select user", sl.Err(err))
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepo) UpdateUserProfile(ctx context.Context, user domain.User) error {
	const op = "Repository.Postgres.UserRepo.UpdateUserProfile"
	logger := r.logger.With(slog.String("op", op))

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if user.FirstName != "" {
		setValues = append(setValues, fmt.Sprintf("first_name=$%d", argID))
		args = append(args, user.FirstName)
		argID++
	}

	if user.LastName != "" {
		setValues = append(setValues, fmt.Sprintf("last_name=$%d", argID))
		args = append(args, user.LastName)
		argID++
	}

	if user.MiddleName != "" {
		setValues = append(setValues, fmt.Sprintf("middle_name=$%d", argID))
		args = append(args, user.MiddleName)
		argID++
	}

	if user.Avatar != "" {
		setValues = append(setValues, fmt.Sprintf("avatar=$%d", argID))
		args = append(args, user.Avatar)
		argID++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(`UPDATE users SET %s WHERE username=$%d`, setQuery, argID)

	args = append(args, user.Username)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		logger.Error("error occurred when update users", sl.Err(err))
		return err
	}

	return nil
}

func (r *UserRepo) ChangeUserPassword(ctx context.Context, userID int, oldPasswordHash, newPasswordHash string) error {
	const op = "Repository.Postgres.UserRepo.UpdateUserProfile"
	logger := r.logger.With(slog.String("op", op))

	query := `UPDATE users
				SET password_hash = $1
					WHERE id = $2 AND password_hash = $3`

	res, err := r.db.Exec(query, newPasswordHash, userID, oldPasswordHash)
	if err != nil {
		logger.Error("error occured when update users", sl.Err(err))
		logger.Debug(fmt.Sprintf("userID: %d"))

		return err
	}

	rowCount, err := res.RowsAffected()
	if err != nil {
		logger.Error("error occured when get RowsAffected", sl.Err(err))
		logger.Debug(fmt.Sprintf("userID: %d", userID))

		return err
	}

	if rowCount == 0 {
		return errors.New("no user found to update")
	}

	// TODO: при смене пароля сбрасывать все сессии

	return nil
}
