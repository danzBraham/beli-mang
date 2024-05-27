package user_exception

import "errors"

var (
	ErrUsernameAlreadyExists   = errors.New("username already exists")
	ErrAdminEmailAlreadyExists = errors.New("admin email already exists")
)
