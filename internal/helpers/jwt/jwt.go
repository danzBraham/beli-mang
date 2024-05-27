package jwt_helper

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var key = []byte(os.Getenv("JWT_SECRET"))

type CustomClaims struct {
	UserId  string
	IsAdmin bool
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
