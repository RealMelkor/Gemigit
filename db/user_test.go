package db

import (
	"testing"
	"strings"
)

func TestGetPublicUser(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	if _, err := GetPublicUser(username + "a"); err == nil {
		t.Fatal("should return user not found")
	}

	user, err := GetPublicUser(username)
	if err != nil {
		t.Fatal(err)
	}

	if user.Name != username {
		t.Fatal("different username")
	}
}

func TestCheckAuth(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	/* Test credential */
	if err := CheckAuth(username, validPassword); err != nil {
		t.Fatal(err)
	}

	if err := CheckAuth(username, validPassword + "a"); err == nil {
		t.Fatal("should return invalid credential")
	}

	if err := CheckAuth(username + "a", validPassword); err == nil {
		t.Fatal("should return invalid credential")
	}
}

func TestChangePassword(t *testing.T) {

	initDB(t)

	username := funcName(t)
	signature := username + "_signature"
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	user, err := FetchUser(username, signature)
	if err != nil {
		t.Fatal(err)
	}

	if err := user.CreateSession(signature); err != nil {
		t.Fatal(err)
	}

	if err := CheckAuth(username, validPassword); err != nil {
		t.Fatal(err)
	}

	if err := user.ChangePassword(invalidPassword, signature); err == nil {
		t.Fatal("password should be invalid")
	}

	err = user.ChangePassword(validPassword + "a", signature + "a")
	if err == nil {
		t.Fatal("signature should be invalid")
	}

	err = user.ChangePassword(validPassword + "a", signature)
	if err != nil {
		t.Fatal(err)
	}

	if err := ChangePassword(username + "a", validPassword); err == nil {
		t.Fatal("username shoudl be invalid")
	}

	if err := CheckAuth(username, validPassword); err == nil {
		t.Fatal("credential should be invalid")
	}

	if err := CheckAuth(username, validPassword + "a"); err != nil {
		t.Fatal(err)
	}

}

func TestRegistration(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	if err := Register(username, validPassword); err == nil {
		t.Fatal("should not be able to register the same username")
	}

	username = strings.ToLower(username)
	if err := Register(username, validPassword); err == nil {
		t.Fatal("should detect case-insentive duplicates")
	}

	if err := Register(invalidUsername, validPassword); err == nil {
		t.Fatal("should not allow invalid username")
	}
}

/*
func TestLogin(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	// Change user password and retest credential
	if err := ChangePassword(username + "a", validPassword + "a");
			err == nil {
		t.Fatal("should return user not found")
	}

	if err := ChangePassword(username, validPassword + "a"); err != nil {
		t.Fatal(err)
	}

	if err := CheckAuth(username, validPassword); err == nil {
		t.Fatal("should return invalid credential")
	}

	if err := CheckAuth(username, validPassword + "a"); err != nil {
		t.Fatal(err)
	}

	// Delete user
	if err := DeleteUser(username + "a"); err == nil {
		t.Fatal("should return user not found")
	}

	if err := DeleteUser(username); err != nil {
		t.Fatal(err)
	}

	if err := CheckAuth(username, validPassword); err == nil {
		t.Fatal("should return invalid credential")
	}
}
*/

