package db

import (
	"testing"
	"gemigit/test"
)

func TestPassword(t *testing.T) {

	test.IsNotNil(t, isPasswordValid(""), "empty passwords are invalid")
	test.IsNotNil(t, isPasswordValid("12345"),
			"short passwords are invalid")
	test.IsNotNil(t, isPasswordValid("1234567890abcdefghjklmnopqrstuvwx"),
			"passwords longer than the length limit are invalid")
	test.IsNil(t, isPasswordValid("123456"))

	hash, err := hashPassword("123456")
	test.IsNil(t, err)

	b := checkPassword("123456", hash)
	test.IsEqual(t, b, true)

	b = checkPassword("1234567", hash)
	test.IsEqual(t, b, false)

}

func TestUsername(t *testing.T) {

	test.IsNotNil(t, isRepoNameValid(""), "empty usernames are invalid")

	test.IsNotNil(t, isUsernameValid("0abc"),
			"name should start with a letter")

	test.IsNotNil(t, isGroupNameValid("iiiiiiiiiiiiiiiiiiiiiiiiI"),
			"name should be shorter than the length limit")

	s := "abc$"
	if isUsernameValid(s) == nil || isGroupNameValid(s) == nil ||
			isRepoNameValid(s) == nil {
		t.Fatal("name should only contain letters and numbers")
	}
	n := "abc0"
	s = "abc_0-"
	test.IsNil(t, isUsernameValid(n))
	test.IsNil(t, isGroupNameValid(s))
	test.IsNil(t, isRepoNameValid(s))
}
