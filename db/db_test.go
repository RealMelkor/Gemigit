package db

import (
	"testing"
	"os"
	"gemigit/test"
)

const validPassword = "pa$$w0rd"
const invalidPassword = "pass"
const invalidUsername = "0user"

var initialized bool
func initDB(t *testing.T) {

	if initialized {
		return
	}

	test.IsNil(t, Init("sqlite3", "test.db", false))

	_, err := db.Exec("DELETE FROM token;")
	test.IsNil(t, err)
	_, err = db.Exec("DELETE FROM user;")
	test.IsNil(t, err)
	_, err = db.Exec("DELETE FROM repo;")
	test.IsNil(t, err)
	_, err = db.Exec("DELETE FROM member;")
	test.IsNil(t, err)
	_, err = db.Exec("DELETE FROM groups;")
	test.IsNil(t, err)
	_, err = db.Exec("DELETE FROM certificate;")
	test.IsNil(t, err)
	_, err = db.Exec("DELETE FROM access;")
	test.IsNil(t, err)

	UpdateTable()
	initialized = true

}

func TestDB(t *testing.T) {

	initDB(t)
	test.IsNil(t, Close())
	test.IsNotNil(t, Init("sqlite3", "/invalid/test.db", false),
			"should be unable to create database")
	test.IsNotNil(t, Init("invalid", "test.db", false),
			"should be unable to open database")
	os.Remove("test.db")
	test.IsNil(t, Init("sqlite3", "test.db", false))
	test.IsNil(t, Close())
	initialized = false
}

func TestUpdateTable(t *testing.T) {

	initDB(t)

	_, err := db.Exec("ALTER TABLE user DROP COLUMN description;")
	test.IsNil(t, err)

	username := test.FuncName(t)
	test.IsNil(t, Register(username, validPassword))

	UpdateTable()
}
