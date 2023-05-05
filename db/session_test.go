package db

import (
	"testing"
)

func TestCreateSession(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)

	if err := user.CreateSession(signature); err == nil {
		t.Fatal("should not be able to add the same signature")
	}
}

func TestDisconnect(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)

	if err := user.Disconnect(signature); err != nil {
		t.Fatal(err)
	}

	if err := user.Disconnect(signature); err == nil {
		t.Fatal("should be already disconnected")
	}
}

func TestGetSessionsCount(t *testing.T) {

	initDB(t)

	user, _, _ := createUserAndSession(t)

	count, err := user.GetSessionsCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("there should be 1 sessions but there is ", count)
	}

	user.ID = -1

	count, err = user.GetSessionsCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("there should be 0 sessions but there is ", count)
	}
}

func TestDisconnectAll(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)

	if err := user.CreateSession(signature + "b"); err != nil {
		t.Fatal(err)
	}

	if err := user.DisconnectAll(signature + "a"); err == nil {
		t.Fatal("should return invalid signature")
	}

	if err := user.DisconnectAll(signature); err != nil {
		t.Fatal(err)
	}
}
