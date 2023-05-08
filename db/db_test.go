package db

import (
	"testing"
	"runtime"
	"strings"
	"os"
	"strconv"
)

func fileAndLine() string {
	_, file, no, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	path := strings.Split(file, "/")
	return path[len(path) - 1] + ":" + strconv.Itoa(no) + ":"
}

func isNil(t *testing.T, err error) {
	if err != nil {
		t.Fatal(fileAndLine(), err)
	}
}

func isNotNil(t *testing.T, err error, message string) {
	if err == nil {
		t.Fatal(fileAndLine(), message)
	}
}

func isEqual(t *testing.T, x interface{}, y interface{}) {
	if x != y {
		t.Fatal(fileAndLine(), x, " != ", y)
	}
}

func isNotEqual(t *testing.T, x interface{}, y interface{}) {
	if x == y {
		t.Fatal(fileAndLine(), x, " != ", y)
	}
}

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

	isNil(t, Init("sqlite3", "test.db", false))

	_, err := db.Exec("DELETE FROM token;")
	isNil(t, err)
	_, err = db.Exec("DELETE FROM user;")
	isNil(t, err)
	_, err = db.Exec("DELETE FROM repo;")
	isNil(t, err)
	_, err = db.Exec("DELETE FROM member;")
	isNil(t, err)
	_, err = db.Exec("DELETE FROM groups;")
	isNil(t, err)
	_, err = db.Exec("DELETE FROM certificate;")
	isNil(t, err)
	_, err = db.Exec("DELETE FROM access;")
	isNil(t, err)

	UpdateTable()
	initialized = true

}

func TestDB(t *testing.T) {

	initDB(t)
	isNil(t, Close())
	isNotNil(t, Init("sqlite3", "/invalid/test.db", false),
			"should be unable to create database")
	isNotNil(t, Init("invalid", "test.db", false),
			"should be unable to open database")
	os.Remove("test.db")
	isNil(t, Init("sqlite3", "test.db", false))
	isNil(t, Close())
	initialized = false
}

func TestUpdateTable(t *testing.T) {

	initDB(t)

	_, err := db.Exec("ALTER TABLE user DROP COLUMN description;")
	isNil(t, err)

	username := funcName(t)
	isNil(t, Register(username, validPassword))

	UpdateTable()
}
