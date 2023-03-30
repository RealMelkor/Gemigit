package db

import (
	"testing"
	"runtime"
	"errors"
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
const invalidUsername = "0user"

var initialized bool
func initDB(t *testing.T) {
	if initialized {
		return
	}
	initialized = true
	os.Remove("test.db")

	if err := Init("sqlite3", "test.db", true); err != nil {
		t.Fatal(err)
	}
}

func TestDB(t *testing.T) {
	initDB(t)
}

func TestRegistration(t *testing.T) {

	initDB(t)

	username := funcName(t)
	if err := Register(username, validPassword); err != nil {
		t.Fatal(err)
	}

	if err := Register(username, validPassword); err == nil {
		t.Fatal(err)
	}

	if err := Register(invalidUsername, validPassword); err == nil {
		t.Fatal(err)
	}
}

func TestSession(t *testing.T) {

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
	if err := AddUserSession(signature, user); err != nil {
		t.Fatal(err)
	}
	if err := AddUserSession(signature, user); err == nil {
		t.Fatal(errors.New(
			"should not be able to add the same signature"))
	}
}
