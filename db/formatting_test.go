package db

import (
	"testing"
)

func TestPassword(t *testing.T) {

	isNotNil(t, isPasswordValid(""), "empty passwords are invalid")
	isNotNil(t, isPasswordValid("12345"), "short passwords are invalid")
	isNotNil(t, isPasswordValid("1234567890abcdefghjklmnopqrstuvwx"),
			"passwords longer than the length limit are invalid")
	isNil(t, isPasswordValid("123456"))

	hash, err := hashPassword("123456")
	isNil(t, err)

	b := checkPassword("123456", hash)
	isEqual(t, b, true)

	b = checkPassword("1234567", hash)
	isEqual(t, b, false)

}

func TestUsername(t *testing.T) {

	isNotNil(t, isRepoNameValid(""), "empty usernames are invalid")

	isNotNil(t, isUsernameValid("0abc"), "name should start with a letter")

	isNotNil(t, isGroupNameValid("iiiiiiiiiiiiiiiiiiiiiiiiI"),
			"name should be shorter than the length limit")

	s := "abc$"
	if isUsernameValid(s) == nil || isGroupNameValid(s) == nil ||
			isRepoNameValid(s) == nil {
		t.Fatal("name should only contain letters and numbers")
	}
	n := "abc0"
	s = "abc_0-"
	isNil(t, isUsernameValid(n))
	isNil(t, isGroupNameValid(s))
	isNil(t, isRepoNameValid(s))
}
