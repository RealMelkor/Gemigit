package db

import (
	"testing"
	"strings"
	"strconv"
	"gemigit/test"
)

const tooLongDescription =
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

var usersCount = 0
func createUserAndSession(t *testing.T) (User, string) {

	usersCount += 1

	username := test.FuncName(t) + strconv.Itoa(usersCount)
	test.IsNil(t, Register(username, validPassword))

	signature := username + "_signature"
	user, err := FetchUser(username, signature)
	test.IsNil(t, err)

	test.IsNil(t, user.CreateSession(signature))

	return user, signature
}

func TestGetUserID(t *testing.T) {

	user, _ := createUserAndSession(t)
	
	id, err := GetUserID(user.Name)
	test.IsNil(t, err)

	test.IsEqual(t, id, user.ID)

	_, err = GetUserID(user.Name + "a")
	test.IsNotNil(t, err, "GetUserID should return user not found")
}

func TestGetUser(t *testing.T) {

	_, signature := createUserAndSession(t)

	user, b := GetUser(signature)
	test.IsEqual(t, b, true)
	test.IsEqual(t, user.Name, user.Name)

	user, b = GetUser(signature + "a")
	test.IsEqual(t, b, false)

	delete(users, signature)

	user, b = GetUser(signature)
	test.IsEqual(t, b, true)
	test.IsEqual(t, user.Name, user.Name)
}

func TestGetPublicUser(t *testing.T) {

	initDB(t)

	username := test.FuncName(t)
	test.IsNil(t, Register(username, validPassword))

	_, err := GetPublicUser(username + "a")
	test.IsNotNil(t, err, "should return user not found")

	user, err := GetPublicUser(username)
	test.IsNil(t, err)
	test.IsEqual(t, user.Name, username)
}

func TestCheckAuth(t *testing.T) {

	initDB(t)

	username := test.FuncName(t)
	test.IsNil(t, Register(username, validPassword))

	/* Test credential */
	test.IsNil(t, CheckAuth(username, validPassword))

	test.IsNotNil(t, CheckAuth(username, validPassword + "a"),
			"should return invalid credential")

	test.IsNotNil(t, CheckAuth(username + "a", validPassword),
			"should return invalid credential")
}

func TestChangePassword(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, CheckAuth(user.Name, validPassword))
	test.IsNotNil(t, user.ChangePassword(invalidPassword, signature), 
			"password should be invalid")
	test.IsNotNil(t,
		user.ChangePassword(validPassword + "a", signature + "a"),
			"signature should be invalid")
	test.IsNil(t, user.ChangePassword(validPassword + "a", signature))
	test.IsNotNil(t, ChangePassword(user.Name+ "a", validPassword),
			"username shoudl be invalid")
	test.IsNotNil(t, CheckAuth(user.Name, validPassword),
			"credential should be invalid")
	test.IsNil(t, CheckAuth(user.Name, validPassword + "a"))
}

func TestRegistration(t *testing.T) {

	initDB(t)

	username := test.FuncName(t)
	test.IsNil(t, Register(username, validPassword))

	test.IsNotNil(t, Register(username, validPassword),
			"should not be able to register the same username")

	username = strings.ToLower(username)
	test.IsNotNil(t, Register(username, validPassword),
			"should detect case-insentive duplicates")

	test.IsNotNil(t, Register(invalidUsername, validPassword),
			"should not allow invalid username")
}

func TestDeleteUser(t *testing.T) {

	initDB(t)

	username := test.FuncName(t)
	test.IsNil(t, Register(username, validPassword))
	test.IsNil(t, CheckAuth(username, validPassword))

	// Delete user
	test.IsNotNil(t, DeleteUser(username + "a"),
			"should return user not found")
	test.IsNil(t, DeleteUser(username))
	test.IsNotNil(t, CheckAuth(username, validPassword),
			"should return invalid credential")

	if err := CheckAuth(username, validPassword); err == nil {
		t.Fatal("should return invalid credential")
	}
}

func TestFetchUser(t *testing.T) {

	initDB(t)

	username := test.FuncName(t)
	test.IsNil(t, Register(username, validPassword))

	signature := username + "_signature"
	user, err := FetchUser(username, signature)
	test.IsNil(t, err)

	test.IsEqual(t, user.Signature, signature)

	_, err = FetchUser(username + "a", signature)
	test.IsNotNil(t, err, "should return user not found")

}

func TestChangeDescription(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNotNil(t, user.ChangeDescription(user.Name, "bad_signature"),
			"should return invalid signature")

	test.IsNil(t, user.ChangeDescription("my description", signature))

	test.IsNotNil(t, user.ChangeDescription(tooLongDescription, signature),
			"should return too long description")
}

func TestUpdateDescription(t *testing.T) {

	initDB(t)

	var u User
	test.IsNotNil(t, u.UpdateDescription(), "should return sql error")

	user, _ := createUserAndSession(t)

	description := "testing"

	_, err := db.Exec("UPDATE user SET description=? WHERE userID=?",
				description, user.ID)
	test.IsNil(t, err)

	test.IsNil(t, user.UpdateDescription())
	test.IsEqual(t, user.Description, description)
}

func TestSetSecret(t *testing.T) {

	initDB(t)

	user, _ := createUserAndSession(t)
	test.IsNil(t, user.SetSecret("secret"))

	user.ID = -1
	test.IsNotNil(t, user.SetSecret("secret"),
			"should return user not found")
}
