package db

import (
	"errors"
	"strconv"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password),
						bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

const (
	passwordMinLen = 6
	passwordMaxLen = 32
	maxNameLen = 24
)

func isPasswordValid(password string) (error) {
	if len(password) == 0 {
		return errors.New("empty password")
	}
	if len(password) < passwordMinLen {
		return errors.New("password too short(minimum " +
				  strconv.Itoa(passwordMinLen) +
				  " characters)")
	}
	if len(password) > passwordMaxLen {
		return errors.New("password too long(maximum " +
				  strconv.Itoa(passwordMaxLen) +
				  " characters)")
	}
	return nil
}

func isNameValid(name string) error {
	if len(name) == 0 {
		return errors.New("empty name")
	}
	if len(name) > maxNameLen {
		return errors.New("name too long")
	}
	if !unicode.IsLetter([]rune(name)[0]) {
		return errors.New("your name must start with a letter")
	}
	return nil
}

func isUsernameValid(name string) error {
	if name == "anon" || name == "root" {
		return errors.New("blacklisted username")
	}
	if err := isNameValid(name); err != nil {
		return err
	}
	for _, c := range name {
		if c > unicode.MaxASCII ||
		   (!unicode.IsLetter(c) && !unicode.IsNumber(c)) {
			return errors.New("your name contains " +
					  "invalid characters")
		}
	}
	return nil
}

func isGroupNameValid(name string) (error) {
	if err := isNameValid(name); err != nil {
		return err
	}
	for _, c := range name {
		if c > unicode.MaxASCII ||
		   (!unicode.IsLetter(c) && !unicode.IsNumber(c) &&
		   c != '-' && c != '_') {
			return errors.New("the group name " +
					  "contains invalid characters")
		}
	}
	return nil
}

func isRepoNameValid(name string) (error) {
	if err := isNameValid(name); err != nil {
		return err
	}
	for _, c := range name {
		if c > unicode.MaxASCII ||
		   (!unicode.IsLetter(c) && !unicode.IsNumber(c) &&
		   c != '-' && c != '_') {
			return errors.New("the repository name " +
					  "contains invalid characters")
		}
	}
	return nil
}
