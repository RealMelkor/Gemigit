package db

import (
	"errors"
	"strconv"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

const (
	passwordMinLen = 6
)

func isPasswordValid(password string) (bool, error) {
	if len(password) == 0 {
		return false, errors.New("empty password")
	}
	if len(password) < passwordMinLen {
		return false, errors.New("password too short(minimum " + strconv.Itoa(passwordMinLen) + " characters)")
	}
	return true, nil
}

func isNameValid(name string) (bool, error) {
	if len(name) == 0 {
		return false, errors.New("empty name")
	}
	if !unicode.IsLetter([]rune(name)[0]) {
		return false, errors.New("your name must start with a letter")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false, errors.New("your name contains invalid characters")
		}
	}
	return true, nil
}

func isRepoNameValid(name string) (bool, error) {
	if len(name) == 0 {
		return false, errors.New("empty name")
	}
	if !unicode.IsLetter([]rune(name)[0]) {
		return false, errors.New("the repository name must start with a letter")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false, errors.New("the repository name contains invalid characters")
		}
	}
	return true, nil
}
