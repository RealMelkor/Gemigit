package db

import (
	"testing"
	"strings"
	"strconv"
)

var usersCount = 0
func createUserAndSession(t *testing.T) (User, string, string) {

	usersCount += 1

	username := funcName(t) + strconv.Itoa(usersCount)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	signature := username + "_signature"
	user, err := FetchUser(username, signature)
	if err != nil {
		t.Fatal(err)
	}

	if err := user.CreateSession(signature); err != nil {
		t.Fatal(err)
	}

	return user, username, signature
}

func TestGetUserID(t *testing.T) {

	user, username, _ := createUserAndSession(t)
	
	id, err := GetUserID(username)
	if err != nil {
		t.Fatal(err)
	}

	if id != user.ID {
		t.Fatal("GetUserID should return the same id as FetchUser")
	}

	_, err = GetUserID(username + "a")
	if err == nil {
		t.Fatal("GetUserID should return user not found")
	}
}

func TestGetUser(t *testing.T) {

	_, username, signature := createUserAndSession(t)

	user, b := GetUser(signature)
	if !b || user.Name != username {
		t.Fatal("GetUser should return the same user")
	}

	user, b = GetUser(signature + "a")
	if b {
		t.Fatal("GetUser should return false")
	}

	delete(users, signature)

	user, b = GetUser(signature)
	if !b || user.Name != username {
		t.Fatal("GetUser should return the same user")
	}
}

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

	user, username, signature := createUserAndSession(t)

	if err := CheckAuth(username, validPassword); err != nil {
		t.Fatal(err)
	}

	if err := user.ChangePassword(invalidPassword, signature); err == nil {
		t.Fatal("password should be invalid")
	}

	err := user.ChangePassword(validPassword + "a", signature + "a")
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

func TestDeleteUser(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	if err := CheckAuth(username, validPassword); err != nil {
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

func TestFetchUser(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	signature := username + "_signature"
	user, err := FetchUser(username, signature)
	if err != nil {
		t.Fatal(err)
	}

	if user.Signature != signature {
		t.Fatal("signature mismatch")
	}

	_, err = FetchUser(username + "a", signature)
	if err == nil {
		t.Fatal("should return user not found")
	}

}

const longString =
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestChangeDescription(t *testing.T) {

	initDB(t)

	user, username, signature := createUserAndSession(t)

	err := user.ChangeDescription(username, "bad_signature")
	if err == nil {
		t.Fatal("should return invalid signature")
	}

	err = user.ChangeDescription("my description", signature)
	if  err != nil {
		t.Fatal(err)
	}

	err = user.ChangeDescription(longString, signature)
	if  err == nil {
		t.Fatal("should return too long description")
	}

}

func TestUpdateDescription(t *testing.T) {

	initDB(t)

	var u User
	if u.UpdateDescription() != nil {
		t.Fatal("should return sql error")
	}

	user, _, _ := createUserAndSession(t)

	description := "testing"

	db.Exec("UPDATE user SET description=? WHERE userID=?",
				description, user.ID)

	user.UpdateDescription()

	if user.Description != description {
		t.Fatal("description value mismatch")
	}
}

func TestSetSecret(t *testing.T) {

	initDB(t)

	user, _, _ := createUserAndSession(t)
	if err := user.SetSecret("secret"); err != nil {
		t.Fatal(err)
	}

	user.ID = -1
	if err := user.SetSecret("secret"); err == nil {
		t.Fatal("should return user not found")
	}
}
