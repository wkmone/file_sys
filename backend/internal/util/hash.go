package util

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("密码长度不能少于 8 个字符")
	}
	if utf8.RuneCountInString(password) > 128 {
		return errors.New("密码长度不能超过 128 个字符")
	}

	hasLetter := false
	hasDigit := false
	for _, r := range password {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			hasLetter = true
		}
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			break
		}
	}
	if !hasLetter {
		return errors.New("密码必须包含至少一个字母")
	}
	if !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}
	return nil
}
