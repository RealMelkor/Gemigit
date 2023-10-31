package db

import (
	"database/sql"
	"log"
	"os"
	"time"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
	_ "github.com/go-sql-driver/mysql"
)

type Repo struct {
	ID		int
	UserID		int
	Username	string
	Name		string
	Date		int
	IsPublic	bool
	Description	string
}

type User struct {
	ID		int
	Name		string
	Description	string
	Registration	int
	Connection	time.Time
	Signature	string
	Secret		string
	SecureGit	bool
}

type Group struct {
	ID		int
	Name		string
	Description	string
}

type Member struct {
	Name	string
	UserID	int
}

type Access struct {
	RepoID		int
	GroupID		int
	UserID		int
	Name		string
	Privilege	int
}

type Token struct {
	Hint			string
	Expiration		int64
	ExpirationFormat	string
	UserID			int
	ID			int
	ReadOnly		bool
}

var unixTime string
var autoincrement string

var db *sql.DB

func Init(dbType string, path string, create bool) error {

	isSqlite := dbType == "sqlite3" || dbType == "sqlite"
	if isSqlite {
		autoincrement = "AUTOINCREMENT"
	} else {
		autoincrement = "AUTO_INCREMENT"
	}

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
		return createTable(db)
	}
	return nil
}

func Close() error {
	return db.Close()
}

func createTable(db *sql.DB) error {

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
		expiration INTEGER NOT NULL,
		readonly INTEGER DEFAULT 0 NOT NULL
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

func printOnSuccess(query string) {
	res, err := db.Exec(query)
	if err == nil {
		if strings.Contains(query, "UPDATE") {
			rows, err := res.RowsAffected()
			if err == nil && rows > 0 {
				log.Println(query)
			}
		} else {
			log.Println(query)
		}
	}
}

/* add missing field to database */
func UpdateTable() {

	/* user table */
	printOnSuccess("CREATE TABLE user;")

	printOnSuccess("ALTER TABLE user ADD " +
		"userID INTEGER NOT NULL PRIMARY KEY " + autoincrement)
	printOnSuccess("ALTER TABLE user ADD name TEXT UNIQUE NOT NULL;")
	printOnSuccess("ALTER TABLE user ADD password TEXT UNIQUE NOT NULL;")

	printOnSuccess(`ALTER TABLE user ADD
			description TEXT DEFAULT "";`)
	printOnSuccess(`UPDATE user SET description=""
			WHERE description IS NULL;`)

	printOnSuccess(`ALTER TABLE user ADD
			secret TEXT DEFAULT "";`)
	printOnSuccess(`UPDATE user SET secret=""
			WHERE secret IS NULL;`)

	printOnSuccess(`ALTER TABLE user ADD
			creation INTEGER NOT NULL;`)
	printOnSuccess(`UPDATE user SET creation=` + unixTime +
			` WHERE creation IS NULL;`)

	printOnSuccess("ALTER TABLE user ADD securegit INTEGER")
	printOnSuccess("UPDATE user SET securegit=0 WHERE securegit IS NULL;")

	printOnSuccess(`CREATE UNIQUE INDEX username_upper ON user (
		UPPER(name), UPPER(name));`)

	/* groups table */
	printOnSuccess("CREATE TABLE groups;")

	printOnSuccess("ALTER TABLE groups ADD " +
		"groupID INTEGER NOT NULL PRIMARY KEY " + autoincrement)
	printOnSuccess("ALTER TABLE groups ADD owner INTEGER NOT NULL;")
	printOnSuccess("ALTER TABLE groups ADD name TEXT UNIQUE NOT NULL;")

	printOnSuccess(`ALTER TABLE groups ADD
			description TEXT DEFAULT "";`)
	printOnSuccess(`UPDATE groups SET description=""
			WHERE description IS NULL;`)

	printOnSuccess(`ALTER TABLE groups ADD
			creation INTEGER NOT NULL;`)
	printOnSuccess(`UPDATE groups SET creation=` + unixTime +
			` WHERE creation IS NULL;`)

	/* member table */
	printOnSuccess("CREATE TABLE member;")
	printOnSuccess("ALTER TABLE member ADD groupID INTEGER NOT NULL;")
	printOnSuccess("ALTER TABLE member ADD userID INTEGER NOT NULL;")

	/* certificate table */
	printOnSuccess("CREATE TABLE certificate;")
	printOnSuccess("ALTER TABLE certificate ADD userID INTEGER NOT NULL;")
	printOnSuccess("ALTER TABLE certificate ADD " +
			"hash TEXT UNIQUE NOT NULL;")

	printOnSuccess(`ALTER TABLE certificate ADD
			creation INTEGER;`)
	printOnSuccess(`UPDATE certificate SET creation=` + unixTime +
			` WHERE creation IS NULL;`)

	/* access table */
	printOnSuccess("CREATE TABLE access;")
	printOnSuccess("ALTER TABLE access ADD repoID INTEGER NOT NULL;")
	printOnSuccess("ALTER TABLE access ADD groupID INTEGER;")
	printOnSuccess("ALTER TABLE access ADD userID INTEGER;")
	printOnSuccess("ALTER TABLE access ADD privilege INTEGER NOT NULL;")

	/* repo table */
	printOnSuccess("CREATE TABLE repo;")
	printOnSuccess("ALTER TABLE repo ADD " +
		"repoID INTEGER NOT NULL PRIMARY KEY " + autoincrement)
	printOnSuccess("ALTER TABLE repo ADD userID INTEGER NOT NULL;")
	printOnSuccess("ALTER TABLE repo ADD name TEXT NOT NULL;")

	printOnSuccess(`ALTER TABLE repo ADD
			description TEXT DEFAULT "";`)
	printOnSuccess(`UPDATE repo SET description=""
			WHERE description IS NULL;`)

	printOnSuccess("ALTER TABLE repo ADD creation INTEGER;")
	printOnSuccess(`UPDATE repo SET creation=` + unixTime +
			` WHERE creation IS NULL;`)

	printOnSuccess("ALTER TABLE repo ADD public INTEGER NOT NULL")
	printOnSuccess("UPDATE repo SET public=0 WHERE public IS NULL;")

	printOnSuccess(`ALTER TABLE repo ADD securegit INTEGER NOT NULL`)
	printOnSuccess(`UPDATE repo SET securegit=0 WHERE securegit IS NULL;`)

	/* token table */
	printOnSuccess("CREATE TABLE token;")
	printOnSuccess("ALTER TABLE token ADD " +
		"tokenID INTEGER NOT NULL PRIMARY KEY " + autoincrement)
	printOnSuccess("ALTER TABLE token ADD userID INTEGER NOT NULL;")
	printOnSuccess("ALTER TABLE token ADD token TEXT NOT NULL;")
	printOnSuccess("ALTER TABLE token ADD hint TEXT NOT NULL;")
	printOnSuccess("ALTER TABLE token ADD expiration INTEGER NOT NULL;")
	printOnSuccess("ALTER TABLE token ADD readonly " +
			"INTEGER DEFAULT 0;")
	printOnSuccess(`UPDATE token SET readonly=0 WHERE readonly IS NULL;`)

}
