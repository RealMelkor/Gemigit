package db

import (
	"testing"
	"strings"
	"strconv"
)

const tooLongDescription =
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

var usersCount = 0
func createUserAndSession(t *testing.T) (User, string) {

	usersCount += 1

	username := funcName(t) + strconv.Itoa(usersCount)
	isNil(t, Register(username, validPassword))

	signature := username + "_signature"
	user, err := FetchUser(username, signature)
	isNil(t, err)

	isNil(t, user.CreateSession(signature))

	return user, signature
}

func TestGetUserID(t *testing.T) {

	user, _ := createUserAndSession(t)
	
	id, err := GetUserID(user.Name)
	isNil(t, err)

	isEqual(t, id, user.ID)

	_, err = GetUserID(user.Name + "a")
	isNotNil(t, err, "GetUserID should return user not found")
}

func TestGetUser(t *testing.T) {

	_, signature := createUserAndSession(t)

	user, b := GetUser(signature)
	isEqual(t, b, true)
	isEqual(t, user.Name, user.Name)

	user, b = GetUser(signature + "a")
	isEqual(t, b, false)

	delete(users, signature)

	user, b = GetUser(signature)
	isEqual(t, b, true)
	isEqual(t, user.Name, user.Name)
}

func TestGetPublicUser(t *testing.T) {

	initDB(t)

	username := funcName(t)
	isNil(t, Register(username, validPassword))

	_, err := GetPublicUser(username + "a")
	isNotNil(t, err, "should return user not found")

	user, err := GetPublicUser(username)
	isNil(t, err)
	isEqual(t, user.Name, username)
}

func TestCheckAuth(t *testing.T) {

	initDB(t)

	username := funcName(t)
	isNil(t, Register(username, validPassword))

	/* Test credential */
	isNil(t, CheckAuth(username, validPassword))

	isNotNil(t, CheckAuth(username, validPassword + "a"),
			"should return invalid credential")

	isNotNil(t, CheckAuth(username + "a", validPassword),
			"should return invalid credential")
}

func TestChangePassword(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	isNil(t, CheckAuth(user.Name, validPassword))
	isNotNil(t, user.ChangePassword(invalidPassword, signature), 
			"password should be invalid")
	isNotNil(t, user.ChangePassword(validPassword + "a", signature + "a"),
			"signature should be invalid")
	isNil(t, user.ChangePassword(validPassword + "a", signature))
	isNotNil(t, ChangePassword(user.Name+ "a", validPassword),
			"username shoudl be invalid")
	isNotNil(t, CheckAuth(user.Name, validPassword),
			"credential should be invalid")
	isNil(t, CheckAuth(user.Name, validPassword + "a"))
}

func TestRegistration(t *testing.T) {

	initDB(t)

	username := funcName(t)
	isNil(t, Register(username, validPassword))

	isNotNil(t, Register(username, validPassword),
			"should not be able to register the same username")

	username = strings.ToLower(username)
	isNotNil(t, Register(username, validPassword),
			"should detect case-insentive duplicates")

	isNotNil(t, Register(invalidUsername, validPassword),
			"should not allow invalid username")
}

func TestDeleteUser(t *testing.T) {

	initDB(t)

	username := funcName(t)
	isNil(t, Register(username, validPassword))
	isNil(t, CheckAuth(username, validPassword))

	// Delete user
	isNotNil(t, DeleteUser(username + "a"), "should return user not found")
	isNil(t, DeleteUser(username))
	isNotNil(t, CheckAuth(username, validPassword),
			"should return invalid credential")

	if err := CheckAuth(username, validPassword); err == nil {
		t.Fatal("should return invalid credential")
	}
}

func TestFetchUser(t *testing.T) {

	initDB(t)

	username := funcName(t)
	isNil(t, Register(username, validPassword))

	signature := username + "_signature"
	user, err := FetchUser(username, signature)
	isNil(t, err)

	isEqual(t, user.Signature, signature)

	_, err = FetchUser(username + "a", signature)
	isNotNil(t, err, "should return user not found")

}

func TestChangeDescription(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	isNotNil(t, user.ChangeDescription(user.Name, "bad_signature"),
			"should return invalid signature")

	isNil(t, user.ChangeDescription("my description", signature))

	isNotNil(t, user.ChangeDescription(tooLongDescription, signature),
			"should return too long description")
}

func TestUpdateDescription(t *testing.T) {

	initDB(t)

	var u User
	isNotNil(t, u.UpdateDescription(), "should return sql error")

	user, _ := createUserAndSession(t)

	description := "testing"

	_, err := db.Exec("UPDATE user SET description=? WHERE userID=?",
				description, user.ID)
	isNil(t, err)

	isNil(t, user.UpdateDescription())
	isEqual(t, user.Description, description)
}

func TestSetSecret(t *testing.T) {

	initDB(t)

	user, _ := createUserAndSession(t)
	isNil(t, user.SetSecret("secret"))

	user.ID = -1
	isNotNil(t, user.SetSecret("secret"), "should return user not found")
}
