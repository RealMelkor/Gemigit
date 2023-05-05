package db

import (
	"testing"
	"runtime"
	"strings"
	"os"
)

func funcName(t *testing.T) string {
	fpcs := make([]uintptr, 1)

	n := runtime.Callers(2, fpcs)
	if n == 0 {
		t.Fatal("function name: no caller")
	}

	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		t.Fatal("function name: caller is nil")
	}

	name := caller.Name()
	return name[strings.LastIndex(name, ".") + 1:]
}

const validPassword = "pa$$w0rd"
const invalidPassword = "pass"
const invalidUsername = "0user"

var initialized bool
func initDB(t *testing.T) {
	if initialized {
		return
	}
	initialized = true
	os.Remove("test.db")

	if err := Init("sqlite3", "test.db", false); err != nil {
		t.Fatal(err)
	}
}

func TestDB(t *testing.T) {
	initDB(t)
	if err := Close(); err != nil {
		t.Fatal(err)
	}

	if err := Init("sqlite3", "/invalid/test.db", false); err == nil {
		t.Fatal("should be unable to create database")
	}

	if err := Init("invalid", "test.db", false); err == nil {
		t.Fatal("should be unable to open database")
	}

	if err := Init("sqlite3", "test.db", false); err != nil {
		t.Fatal(err)
	}

}

func TestUpdateTable(t *testing.T) {

	initDB(t)

	_, err := db.Exec("ALTER TABLE user DROP COLUMN description;")
	if err != nil {
		t.Fatal(err)
	}


	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	UpdateTable()
}

/*
func TestSession(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	signature := username + "_signature"

	user, err := FetchUser(username + "a", signature)
	if err == nil {
		t.Fatal("should return user not found", user)
	}

	user, err = FetchUser(username, signature)
	if err != nil {
		t.Fatal(err)
	}

	if err := user.CreateSession(signature); err != nil {
		t.Fatal(err)
	}

	if err := user.CreateSession(signature); err == nil {
		t.Fatal("should not be able to add the same signature")
	}

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
*/

/*
func TestVerifySignature(t *testing.T) {

	first := funcName(t) + "0"
	second := funcName(t) + "1"
	first_signature := first + "_signature"
	second_signature := second + "_signature"

	if err := Register(first, validPassword); err != nil {
		t.Fatal(err)
	}

	if err := Register(second, validPassword); err != nil {
		t.Fatal(err)
	}

	user, err := FetchUser(first, first_signature)
	if err != nil {
		t.Fatal(err)
	}

	other, err := FetchUser(second, second_signature)
	if err != nil {
		t.Fatal(err)
	}

	if err := user.CreateSession(first_signature); err != nil {
		t.Fatal(err)
	}

	if err := other.CreateSession(second_signature); err != nil {
		t.Fatal(err)
	}

	if err := user.VerifySignature(first_signature); err != nil {
		t.Fatal(err)
	}

	if err := user.VerifySignature(first_signature + "a"); err == nil {
		t.Fatal("the signature should be wrong")
	}

	if err := user.VerifySignature(second_signature); err == nil {
		t.Fatal("the signature should not match")
	}
}
*/
