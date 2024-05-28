package jwt_helper

import (
	"fmt"
	"os"
	"time"

	auth_exception "github.com/danzBraham/beli-mang/internal/exceptions/auth"
	"github.com/golang-jwt/jwt/v5"
)

var key = []byte(os.Getenv("JWT_SECRET"))

type CustomClaims struct {
	UserId  string `json:"userId"`
	IsAdmin bool   `json:"isAdmin"`
	jwt.RegisteredClaims
}

func GenerateToken(ttl time.Duration, userId string, isAdmin bool) (string, error) {
	now := time.Now()
	expiry := now.Add(ttl)

	claims := &CustomClaims{
		UserId:  userId,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

type JWTPayload struct {
	UserId  string
	IsAdmin bool
}

func VerifyToken(tokenString string) (*JWTPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return key, nil
	})
	if token == nil {
		return nil, auth_exception.ErrMissingToken
	}
	if err != nil {
		return nil, auth_exception.ErrInvalidToken
	}
	if !token.Valid {
		return nil, auth_exception.ErrInvalidToken
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || claims == nil {
		return nil, auth_exception.ErrUnknownClaims
	}

	return &JWTPayload{
		UserId:  claims.UserId,
		IsAdmin: claims.IsAdmin,
	}, nil
}
