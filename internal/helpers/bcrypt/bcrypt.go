package bcrypt_helper

import (
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	salt, err := strconv.Atoi(os.Getenv("BCRYPT_SALT"))
	if err != nil {
		return "", err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), salt)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
