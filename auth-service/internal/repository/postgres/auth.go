package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/shamank/edutour-backend/auth-service/internal/domain"
	"github.com/shamank/edutour-backend/auth-service/pkg/logger/sl"
	"log/slog"
)

const (
	tokenTypeEmail    = 1
	tokenTypePassword = 2
)

type AuthRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAuthRepo(db *sql.DB, logger *slog.Logger) *AuthRepo {
	return &AuthRepo{db: db, logger: logger}
}

func (r *AuthRepo) Create(ctx context.Context, user domain.User, confirmToken string, expireAt int64) error {
	const op = "Repository.Postgres.AuthRepo.Create"

	logger := r.logger.With(slog.String("op", op))

	insertUserQuery := `INSERT INTO USERS (username, email, password_hash) 
			VALUES ($1, $2, $3) RETURNING id`

	tx, err := r.db.Begin()
	if err != nil {
		logger.Error("fail create r.db.Begin()!", sl.Err(err))
		return err
	}

	var id int
	row := tx.QueryRow(insertUserQuery, user.Username, user.Email, user.PasswordHash)
	if err := row.Scan(&id); err != nil {

		logger.Error("error occurred when insert new user", sl.Err(err))

		tx.Rollback()
		return err
	}

	insertUserTokensQuery := `INSERT INTO USER_TOKENS (user_id, token_type, token_value, expire_at)
				VALUES ($1, $2, $3, to_timestamp($4))`

	_, err = tx.Exec(insertUserTokensQuery, id, tokenTypeEmail, confirmToken, expireAt)
	if err != nil {

		logger.Error("error occurred when insert new user_token", sl.Err(err))

		tx.Rollback()
		return err
	}

	//logger.Debug("created new user:")

	return tx.Commit()
}

func (r *AuthRepo) ConfirmUser(ctx context.Context, confirmToken string) error {
	const op = "Repository.Postgres.AuthRepo.ConfirmUser"
	logger := r.logger.With(slog.String("op", op))

	//query1 := `UPDATE USER_TOKENS
	//			SET black_list = True
	//			WHERE token_type = $1 AND token_value = $2 AND black_list = FALSE AND expire_at > CURRENT_TIMESTAMP
	//			RETURNING user_id`

	// TODO: сделать так, чтобы при expire_at удалялся аккаунт и не проходила проверка...
	query1 := `UPDATE USER_TOKENS
				SET black_list = True
				WHERE token_type = $1 AND token_value = $2 AND black_list = FALSE
				RETURNING user_id`

	query2 := `UPDATE USERS
				SET is_confirm = True
				WHERE id = $1`

	tx, err := r.db.Begin()
	if err != nil {
		logger.Error("fail create r.db.Begin()!", sl.Err(err))
		return err
	}

	var Id int

	row := tx.QueryRow(query1, tokenTypeEmail, confirmToken)
	if err := row.Scan(&Id); err != nil {
		logger.Error("error occurred when inserting in user_tokens", sl.Err(err))

		tx.Rollback()
		return err
	}

	logger.Debug(fmt.Sprintf("confirming user with id: %d", Id))

	_, err = tx.Exec(query2, Id)
	if err != nil {
		logger.Error("error occurred when update user table", sl.Err(err))

		tx.Rollback()
		return err
	}

	tx.Commit()

	return err
}

func (r *AuthRepo) SetTokenResetPassword(ctx context.Context, email string, token string, expireAt int64) error {
	const op = "Repository.Postgres.AuthRepo.SetTokenResetPassword"
	logger := r.logger.With(slog.String("op", op))

	tx, err := r.db.Begin()
	if err != nil {
		logger.Error("fail create r.db.Begin()!", sl.Err(err))
		return err
	}
	query1 := `SELECT u.ID from USERS u where email = $1 AND is_confirm = TRUE`

	var userID int

	row := tx.QueryRow(query1, email)
	if err := row.Scan(&userID); err != nil {
		logger.Error("error occurred when select user", sl.Err(err))

		tx.Rollback()
		return err
	}

	query2 := `INSERT INTO user_tokens(user_id, token_type, token_value, expire_at)
				VALUES($1, $2, $3, to_timestamp($4))`

	_, err = tx.Exec(query2, userID, tokenTypePassword, token, expireAt)
	if err != nil {
		logger.Error("error occurred when insert into user_tokens", sl.Err(err))

		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *AuthRepo) ConfirmResetPassword(ctx context.Context, token string, passwordHash string) error {
	const op = "Repository.Postgres.AuthRepo.ConfirmResetPassword"
	logger := r.logger.With(slog.String("op", op))

	tx, err := r.db.Begin()
	if err != nil {
		logger.Error("fail create r.db.Begin()!", sl.Err(err))
		return err
	}

	// TODO: при смене пароля сбрасывать все сессии

	//query1 := `SELECT user_id FROM user_tokens
	//			WHERE token_type = $1 AND token_value = $2 AND expire_at < CURRENT_TIMESTAMP`

	query1 := `SELECT user_id FROM user_tokens
				WHERE token_type = $1 AND token_value = $2 AND expire_at > CURRENT_TIMESTAMP`

	var userID int

	row := tx.QueryRow(query1, tokenTypePassword, token)
	if err := row.Scan(&userID); err != nil {
		logger.Error("error occurred when select from user_tokens", sl.Err(err))

		tx.Rollback()
		return err
	}

	query2 := `UPDATE USERS
			SET password_hash = $1 WHERE id = $2`
	_, err = tx.Exec(query2, passwordHash, userID)
	if err != nil {
		logger.Error("error occurred when update users", sl.Err(err))
		tx.Rollback()
		return err
	}

	query3 := `UPDATE USER_TOKENS
				SET black_list = True WHERE token_type = $1 AND token_value = $2`
	_, err = tx.Exec(query3, tokenTypePassword, token)
	if err != nil {
		logger.Error("error occurred when update user_tokens", sl.Err(err))
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (r *AuthRepo) GetByCredentials(ctx context.Context, email string, passwordHash string) (domain.User, error) {
	const op = "Repository.Postgres.AuthRepo.GetByCredentials"
	logger := r.logger.With(slog.String("op", op))

	var user domain.User

	query := `SELECT u.id, u.username, u.email, u.role_id, r.name
				FROM USERS u
				INNER JOIN ROLE_TYPES r on u.role_id = r.id
				WHERE u.email = $1 AND u.password_hash = $2`

	err := r.db.QueryRow(query, email, passwordHash).Scan(&user.ID,
		&user.Username,
		&user.Email,
		&user.Role.ID,
		&user.Role.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		logger.Error(fmt.Sprintf("error occurred when select from users: %s, %s",
			email, passwordHash),
			sl.Err(err))
	}

	return user, err
}

func (r *AuthRepo) GetByUsername(ctx context.Context, username string, passwordHash string) (domain.User, error) {
	const op = "Repository.Postgres.AuthRepo.GetByUsername"
	logger := r.logger.With(slog.String("op", op))

	var user domain.User

	query := `SELECT u.id, u.username, u.email, u.role_id, r.name
				FROM USERS u
				INNER JOIN ROLE_TYPES r on u.role_id = r.id
				WHERE u.username = $1 AND u.password_hash = $2`

	err := r.db.QueryRow(query, username, passwordHash).Scan(&user.ID,
		&user.Username,
		&user.Email,
		&user.Role.ID,
		&user.Role.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		logger.Error("error occurred when select from users: ", sl.Err(err))
		return domain.User{}, err
	}

	return user, nil
}

func (r *AuthRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (domain.User, error) {
	const op = "Repository.Postgres.AuthRepo.GetByRefreshToken"
	logger := r.logger.With(slog.String("op", op))

	var user domain.User
	var tokenID int
	query := `SELECT u.id, u.username, u.email, u.role_id, r.name, t.id
				FROM USERS u
				INNER JOIN ROLE_TYPES r on r.id = u.role_id
				INNER JOIN REFRESH_TOKENS t on t.user_id = u.id
				WHERE t.refresh_token = $1 AND t.expire_at > CURRENT_TIMESTAMP AND NOT t.black_list`

	err := r.db.QueryRow(query, refreshToken).Scan(&user.ID,
		&user.Username,
		&user.Email,
		&user.Role.ID,
		&user.Role.Name,
		&tokenID)
	if err != nil {
		logger.Error("error occurred when select from users", sl.Err(err))
		return domain.User{}, err
	}

	query2 := `UPDATE REFRESH_TOKENS
				SET black_list = true
				WHERE id = $1`
	_, err = r.db.Exec(query2, tokenID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		logger.Error("error occurred when update refresh tokens", sl.Err(err))
		return domain.User{}, err
	}

	return user, nil
}

func (r *AuthRepo) SetRefreshToken(ctx context.Context, userID int, refreshInput domain.RefreshToken) error {
	const op = "Repository.Postgres.AuthRepo.SetRefreshToken"
	logger := r.logger.With(slog.String("op", op))

	query := `INSERT INTO REFRESH_TOKENS (user_id, refresh_token, expire_at) VALUES ($1, $2, to_timestamp($3))`

	_, err := r.db.Exec(query, userID, refreshInput.RefreshToken, int(refreshInput.ExpiresAt))

	if err != nil {
		logger.Error("error occurred when insert into refresh_tokens", sl.Err(err))
		return err
	}

	return nil
}

func (r *AuthRepo) Verify(ctx context.Context, userID int) error {
	return nil
}

func (r *AuthRepo) GetFullUserInfo(ctx context.Context, userID int) (domain.User, error) {
	const op = "Repository.Postgres.AuthRepo.SetRefreshToken"
	logger := r.logger.With(slog.String("op", op))

	var u domain.User

	query := `SELECT u.id, u.username, u.email, COALESCE(u.phone, '') as phone, 
				COALESCE(u.avatar, ''),  COALESCE(u.first_name, '') as first_name,
       COALESCE(u.last_name, '') as last_name, COALESCE(u.middle_name, '') as middle_name,
				u.created_at, r.id, r.name
				FROM USERS u
				INNER JOIN ROLE_TYPES r on r.id = u.role_id
				WHERE u.id = $1`
	err := r.db.QueryRow(query, userID).Scan(&u.ID,
		&u.Username,
		&u.Email,
		&u.Phone,
		&u.Avatar,
		&u.FirstName,
		&u.LastName,
		&u.MiddleName,
		&u.CreatedAt,
		&u.Role.ID,
		&u.Role.Name)
	if err != nil {
		logger.Error("error occurred when insert users", sl.Err(err))
		return domain.User{}, err
	}
	return u, nil
}
