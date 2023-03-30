package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
	_ "github.com/go-sql-driver/mysql"
)

type Repo struct {
	RepoID      int
	UserID      int
	Username    string
	Name        string
	Date        int
	IsPublic    bool
	Description string
}

type User struct {
	ID           int
	Name         string
	Description  string
	Registration int
	Connection   time.Time
	Signature    string
	Secret       string
	SecureGit    bool
}

type Group struct {
	GroupID     int
	Name        string
	Description string
}

type Member struct {
	Name	string
	UserID	int
}

type Access struct {
	RepoID int
	GroupID int
	UserID	int
	Name	string
	Privilege int
}

type Token struct {
	Hint string
	Expiration int64
	ExpirationFormat string
	UserID	int
	ID int
}

var unixTime string

var db *sql.DB

func Init(dbType string, path string, create bool) error {

	isSqlite := dbType == "sqlite3" || dbType == "sqlite"

	if !create && isSqlite {
		file, err := os.Open(path)
		if os.IsNotExist(err) {
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			file.Close()
			log.Println("Creating database " + path)
			create = true
		} else {
			file.Close()
			log.Println("Loading database " + path)
		}
	}

	var err error
	db, err = sql.Open(dbType, path)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	unixTime = "UNIX_TIMESTAMP()"
	if isSqlite {
		unixTime = "strftime('%s', 'now')"
	}
	if create {
		return createTable(db, isSqlite)
	}
	return nil
}

func Close() error {
	return db.Close()
}

func createTable(db *sql.DB, isSqlite bool) error {
	autoincrement := "AUTO_INCREMENT"

	if isSqlite {
		autoincrement = "AUTOINCREMENT"
	}

	userTable := `CREATE TABLE user (
		userID INTEGER NOT NULL PRIMARY KEY ` + autoincrement + `,
		name TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		description TEXT DEFAULT "" NOT NULL,
		secret TEXT DEFAULT "" NOT NULL,
		creation INTEGER NOT NULL,
		securegit INTEGER DEFAULT 0 NOT NULL
	);`

	groupTable := `CREATE TABLE groups (
		groupID INTEGER NOT NULL PRIMARY KEY ` + autoincrement + `,
		owner INTEGER NOT NULL,
		name TEXT UNIQUE NOT NULL,
		description TEXT DEFAULT "",
		creation INTEGER NOT NULL
	);`

	memberTable := `CREATE TABLE member (
		groupID INTEGER NOT NULL,
		userID INTEGER NOT NULL
	);`

	certificateTable := `CREATE TABLE certificate (
		userID INTEGER NOT NULL,
		hash TEXT UNIQUE NOT NULL,
		creation INTEGER NOT NULL
	);`

	accessTable := `CREATE TABLE access (
		repoID INTEGER NOT NULL,
		groupID INTEGER,
		userID INTEGER,
		privilege INTEGER NOT NULL
	);`

	repoTable := `CREATE TABLE repo (
		repoID INTEGER NOT NULL PRIMARY KEY ` + autoincrement + `,
		userID INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT DEFAULT "",
		creation INTEGER NOT NULL,
		public INTEGER DEFAULT 0,
		securegit INTEGER DEFAULT 0 NOT NULL
	);`

	tokenTable := `CREATE TABLE token (
		tokenID INTEGER NOT NULL PRIMARY KEY ` + autoincrement + `,
		userID INTEGER NOT NULL,
		token TEXT NOT NULL,
		hint TEXT NOT NULL,
		expiration INTEGER NOT NULL
	);`

	userConstraint := `CREATE UNIQUE INDEX username_upper ON user (
		UPPER(name), UPPER(name)
	);`

	_, err := db.Exec(userTable)
	if err != nil {
		return err
	}
	log.Println("Users table created")

	_, err = db.Exec(groupTable)
	if err != nil {
		return err
	}
	log.Println("Groups table created")

	_, err = db.Exec(memberTable)
	if err != nil {
		return err
	}
	log.Println("Members table created")

	_, err = db.Exec(certificateTable)
	if err != nil {
		return err
	}
	log.Println("Certificates table created")

	_, err = db.Exec(accessTable)
	if err != nil {
		return err
	}
	log.Println("Access table created")

	_, err = db.Exec(repoTable)
	if err != nil {
		return err
	}
	log.Println("Repositories table created")

	_, err = db.Exec(tokenTable)
	if err != nil {
		return err
	}
	log.Println("Tokens table created")

        _, err = db.Exec(userConstraint)
        if err != nil {
                return err
        }
        log.Println("Users constraint created")

	return nil
}
