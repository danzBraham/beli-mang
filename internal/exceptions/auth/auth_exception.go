package auth_exception

import "errors"

var (
	ErrMissingToken  = errors.New("missing token")
	ErrInvalidToken  = errors.New("invalid token")
	ErrUnknownClaims = errors.New("unknown claims type")
)
