package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword nhận vào mật khẩu thô, trả về chuỗi đã được băm
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
