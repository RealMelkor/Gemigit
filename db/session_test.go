package db

import (
	"testing"
	"gemigit/test"
)

func TestCreateSession(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	test.IsNotNil(t, user.CreateSession(signature),
			"should not be able to add the same signature")
}

func TestDisconnect(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, user.Disconnect(signature))
	test.IsNotNil(t, user.Disconnect(signature),
			"should be already disconnected")
}

func TestGetSessionsCount(t *testing.T) {

	initDB(t)

	user, _ := createUserAndSession(t)

	count, err := user.GetSessionsCount()
	test.IsNil(t, err)
	test.IsEqual(t, count, 1)

	test.IsNil(t, user.CreateSession("new_signature"))

	count, err = user.GetSessionsCount()
	test.IsNil(t, err)
	test.IsEqual(t, count, 2)

	user.ID = -1

	count, err = user.GetSessionsCount()
	test.IsNil(t, err)
	test.IsEqual(t, count, 0)
}

func TestDisconnectAll(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	test.IsNil(t, user.CreateSession(signature + "b"))
	test.IsNotNil(t, user.DisconnectAll(signature + "a"),
			"should return invalid signature")
	test.IsNil(t, user.DisconnectAll(signature))
}
