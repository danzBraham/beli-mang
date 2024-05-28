package user_exception

import "errors"

var (
	ErrUsernameAlreadyExists   = errors.New("username already exists")
	ErrAdminEmailAlreadyExists = errors.New("admin email already exists")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidPassword         = errors.New("invalid password")
)
