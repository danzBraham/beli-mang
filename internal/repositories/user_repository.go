package repositories

import (
	"context"
	"errors"

	user_entity "github.com/danzBraham/beli-mang/internal/entities/user"
	user_exception "github.com/danzBraham/beli-mang/internal/exceptions/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	VerifyUsername(ctx context.Context, username string) (bool, error)
	VerifyAdminEmail(ctx context.Context, email string) (bool, error)
	CreateAdminUser(ctx context.Context, user *user_entity.User) error
	GetUserByUsername(ctx context.Context, username string) (*user_entity.User, error)
}

type UserRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &UserRepositoryImpl{DB: db}
}

func (r *UserRepositoryImpl) VerifyUsername(ctx context.Context, username string) (bool, error) {
	var one int
	query := `SELECT 1 FROM users WHERE username = $1`
	err := r.DB.QueryRow(ctx, query, username).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *UserRepositoryImpl) VerifyAdminEmail(ctx context.Context, email string) (bool, error) {
	var one int
	query := `SELECT 1 FROM users WHERE email = $1 AND is_admin = true`
	err := r.DB.QueryRow(ctx, query, email).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *UserRepositoryImpl) CreateAdminUser(ctx context.Context, user *user_entity.User) error {
	query := `INSERT INTO users (id, username, password, email, is_admin)
						VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(ctx, query, &user.Id, &user.Username, &user.Password, &user.Email, &user.IsAdmin)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepositoryImpl) GetUserByUsername(ctx context.Context, username string) (*user_entity.User, error) {
	var user user_entity.User
	query := `SELECT id, username, password, email, is_admin FROM users WHERE username = $1`
	err := r.DB.QueryRow(ctx, query, username).Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, user_exception.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
