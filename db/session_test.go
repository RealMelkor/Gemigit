package db

import (
	"testing"
)

func TestCreateSession(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	isNotNil(t, user.CreateSession(signature),
			"should not be able to add the same signature")
}

func TestDisconnect(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	isNil(t, user.Disconnect(signature))
	isNotNil(t, user.Disconnect(signature),
			"should be already disconnected")
}

func TestGetSessionsCount(t *testing.T) {

	initDB(t)

	user, _ := createUserAndSession(t)

	count, err := user.GetSessionsCount()
	isNil(t, err)
	isEqual(t, count, 1)

	isNil(t, user.CreateSession("new_signature"))

	count, err = user.GetSessionsCount()
	isNil(t, err)
	isEqual(t, count, 2)

	user.ID = -1

	count, err = user.GetSessionsCount()
	isNil(t, err)
	isEqual(t, count, 0)
}

func TestDisconnectAll(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	isNil(t, user.CreateSession(signature + "b"))
	isNotNil(t, user.DisconnectAll(signature + "a"),
			"should return invalid signature")
	isNil(t, user.DisconnectAll(signature))
}
