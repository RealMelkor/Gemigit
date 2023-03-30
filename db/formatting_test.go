package db

import (
	"testing"
)

func TestPassword(t *testing.T) {

	if isPasswordValid("") == nil {
		t.Fatal("empty passwords are invalid")
	}

	if isPasswordValid("12345") == nil {
		t.Fatalf("password shorter than %d are invalid",
				passwordMinLen)
	}

	if isPasswordValid("1234567890abcdefghjklmnopqrstuvwx") == nil {
		t.Fatalf("passwords longer than %d are invalid",
				passwordMaxLen)
	}

	if isPasswordValid("123456") != nil {
		t.Fatal("password should be valid")
	}

	hash, err := hashPassword("123456")
	if err != nil {
		t.Fatal(err)
	}

	b := checkPassword("123456", hash)
	if !b {
		t.Fatal("password should be valid")
	}

	b = checkPassword("1234567", hash)
	if b {
		t.Fatal("password should be invalid")
	}
}

func TestUsername(t *testing.T) {

	s := ""
	if isRepoNameValid(s) == nil {
		t.Fatal("empty usernames are invalid")
	}

	s = "0abc"
	if isUsernameValid(s) == nil {
		t.Fatal("name should start with a letter")
	}

	s = "iiiiiiiiiiiiiiiiiiiiiiiiI"
	if isGroupNameValid(s) == nil {
		t.Fatalf("name should contain less than %d characters",
				maxNameLen)
	}

	s = "abc$"
	if isUsernameValid(s) == nil || isGroupNameValid(s) == nil ||
			isRepoNameValid(s) == nil {
		t.Fatal("name should only contain letters and numbers")
	}
	n := "abc0"
	s = "abc_0-"
	if isUsernameValid(n) != nil || isGroupNameValid(s) != nil ||
			isRepoNameValid(s) != nil {
		t.Fatal("name should be valid")
	}
}
