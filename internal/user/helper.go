package user

import (
	"golang.org/x/crypto/bcrypt"
)

func verifyPassword(providedPassword, storedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword))
	return err == nil
}
