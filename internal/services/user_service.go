package services

import (
	"context"
	"time"

	user_entity "github.com/danzBraham/beli-mang/internal/entities/user"
	user_exception "github.com/danzBraham/beli-mang/internal/exceptions/user"
	bcrypt_helper "github.com/danzBraham/beli-mang/internal/helpers/bcrypt"
	jwt_helper "github.com/danzBraham/beli-mang/internal/helpers/jwt"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type UserService interface {
	RegisterAdminUser(ctx context.Context, payload *user_entity.RegisterUserRequest) (*user_entity.RegisterUserResponse, error)
	LoginAdminUser(ctx context.Context, payload *user_entity.LoginUserRequest) (*user_entity.LoginUserResponse, error)
	RegisterUser(ctx context.Context, payload *user_entity.RegisterUserRequest) (*user_entity.RegisterUserResponse, error)
	LoginUser(ctx context.Context, payload *user_entity.LoginUserRequest) (*user_entity.LoginUserResponse, error)
}

type UserServiceImpl struct {
	Repository repositories.UserRepository
}

func NewUserService(repository repositories.UserRepository) UserService {
	return &UserServiceImpl{Repository: repository}
}

func (s *UserServiceImpl) RegisterAdminUser(ctx context.Context, payload *user_entity.RegisterUserRequest) (*user_entity.RegisterUserResponse, error) {
	isUsernameExists, err := s.Repository.VerifyUsername(ctx, payload.Username)
	if err != nil {
		return nil, err
	}
	if isUsernameExists {
		return nil, user_exception.ErrUsernameAlreadyExists
	}

	isAdminEmailExists, err := s.Repository.VerifyAdminEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}
	if isAdminEmailExists {
		return nil, user_exception.ErrAdminEmailAlreadyExists
	}

	hashedPassword, err := bcrypt_helper.HashPassword(payload.Password)
	if err != nil {
		return nil, err
	}

	user := &user_entity.User{
		Id:       ulid.Make().String(),
		Username: payload.Username,
		Password: hashedPassword,
		Email:    payload.Email,
		IsAdmin:  true,
	}

	err = s.Repository.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := jwt_helper.GenerateToken(2*time.Hour, user.Id, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &user_entity.RegisterUserResponse{
		Token: token,
	}, nil
}

func (s *UserServiceImpl) LoginAdminUser(ctx context.Context, payload *user_entity.LoginUserRequest) (*user_entity.LoginUserResponse, error) {
	user, err := s.Repository.GetAdminUserByUsername(ctx, payload.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt_helper.VerifyPassword(user.Password, payload.Password)
	if err != nil {
		return nil, user_exception.ErrInvalidPassword
	}

	token, err := jwt_helper.GenerateToken(2*time.Hour, user.Id, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &user_entity.LoginUserResponse{
		Token: token,
	}, nil
}

func (s *UserServiceImpl) RegisterUser(ctx context.Context, payload *user_entity.RegisterUserRequest) (*user_entity.RegisterUserResponse, error) {
	isUsernameExists, err := s.Repository.VerifyUsername(ctx, payload.Username)
	if err != nil {
		return nil, err
	}
	if isUsernameExists {
		return nil, user_exception.ErrUsernameAlreadyExists
	}

	isUserEmailExists, err := s.Repository.VerifyUserEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}
	if isUserEmailExists {
		return nil, user_exception.ErrUserEmailAlreadyExists
	}

	hashedPassword, err := bcrypt_helper.HashPassword(payload.Password)
	if err != nil {
		return nil, err
	}

	user := &user_entity.User{
		Id:       ulid.Make().String(),
		Username: payload.Username,
		Password: hashedPassword,
		Email:    payload.Email,
		IsAdmin:  false,
	}

	err = s.Repository.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := jwt_helper.GenerateToken(2*time.Hour, user.Id, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &user_entity.RegisterUserResponse{
		Token: token,
	}, nil
}

func (s *UserServiceImpl) LoginUser(ctx context.Context, payload *user_entity.LoginUserRequest) (*user_entity.LoginUserResponse, error) {
	user, err := s.Repository.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt_helper.VerifyPassword(user.Password, payload.Password)
	if err != nil {
		return nil, user_exception.ErrInvalidPassword
	}

	token, err := jwt_helper.GenerateToken(2*time.Hour, user.Id, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &user_entity.LoginUserResponse{
		Token: token,
	}, nil
}
