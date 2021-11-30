package db

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

/*
type users struct {
	username
}*/
var users = make(map[string]string)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

const (
	passwordMinLen = 6
)

func isPasswordValid(password string) (bool, error) {
	if len(password) == 0 {
		return false, errors.New("empty password")
	}
	if len(password) < passwordMinLen {
		return false, errors.New("password too short(minimum " + strconv.Itoa(passwordMinLen) + " characters)")
	}
	return true, nil
}

func isNameValid(name string) (bool, error) {
	if len(name) == 0 {
		return false, errors.New("empty name")
	}
	if !unicode.IsLetter([]rune(name)[0]) {
		return false, errors.New("your name must start with a letter")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false, errors.New("your name contains invalid characters")
		}
	}
	return true, nil
}

func isRepoNameValid(name string) (bool, error) {
	if len(name) == 0 {
		return false, errors.New("empty name")
	}
	if !unicode.IsLetter([]rune(name)[0]) {
		return false, errors.New("the repository name must start with a letter")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false, errors.New("the repository name contains invalid characters")
		}
	}
	return true, nil
}

func userAlreadyExist(username string) (bool, error) {
	rows, err := db.Query("select * from user WHERE UPPER(name) LIKE UPPER(?)", username)
	if err != nil {
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func repoAlreadyExist(username string, repo string) (bool, error) {
	rows, err := db.Query("SELECT * FROM repo a INNER JOIN user b ON a.userID=b.userID WHERE UPPER(a.name) LIKE UPPER(?) AND UPPER(b.name) LIKE UPPER(?)", username, repo)
	if err != nil {
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func getUserID(username string) (int, error) {
	rows, err := db.Query("select userID from user WHERE UPPER(name) LIKE UPPER(?)", username)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if rows.Next() {
		var uID int
		err = rows.Scan(&uID)
		if err != nil {
			return -1, err
		}
		return uID, nil
	}
	return -1, errors.New("No user with username " + username)
}

var db *sql.DB

func Init(path string) error {

	new := false
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		file.Close()
		log.Println("Creating database " + path)
		new = true
	} else {
		file.Close()
		log.Println("Loading database " + path)
	}

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	if new {
		return createTable(db)
	}
	return nil
}

func Close() error {
	return db.Close()
}

func createTable(db *sql.DB) error {
	createUserTable := `CREATE TABLE user (
		"userID" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"name" TEXT,
		"password" TEXT,
		"creation" INTEGER		
	  );`

	createRepoTable := `CREATE TABLE repo (
		"repoID" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"userID" integer,	
		"name" TEXT,
		"creation" INTEGER,
		"public" INTEGER	
	);`

	statement, err := db.Prepare(createUserTable)
	if err != nil {
		return err
	}
	if _, err = statement.Exec(); err != nil {
		return err
	}
	log.Println("Users table created")
	statement, err = db.Prepare(createRepoTable)
	if err != nil {
		return err
	}
	if _, err = statement.Exec(); err != nil {
		return err
	}
	log.Println("Repositories table created")
	return nil
}

func CheckAuth(username string, password string) (bool, error) {
	rows, err := db.Query("select name, password from user WHERE UPPER(name) LIKE UPPER(?)", username)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var dPassword string
		var dName string
		err = rows.Scan(&dName, &dPassword)
		if err != nil {
			return false, err
		}
		if checkPassword(password, dPassword) {
			return true, nil
		}
	}
	return false, nil
}

func Login(username string, password string, signature string) (bool, error) {
	rows, err := db.Query("select name, password, userID from user WHERE UPPER(name) LIKE UPPER(?)", username)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var dPassword string
		var dName string
		var dID int
		err = rows.Scan(&dName, &dPassword, &dID)
		if err != nil {
			return false, err
		}
		if checkPassword(password, dPassword) {
			users[signature] = username
			return true, nil
		}
	}
	return false, nil
}

func Register(username string, password string) error {

	if isValid, err := isPasswordValid(password); !isValid {
		return err
	}

	if isValid, err := isNameValid(username); !isValid {
		return err
	}

	if exist, err := userAlreadyExist(username); exist || err != nil {
		if err != nil {
			return err
		}
		return errors.New("this name is already taken")
	}

	statement, err := db.Prepare("insert into user(name,password,creation) VALUES(?,?,strftime('%s', 'now'));")
	if err != nil {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = statement.Exec(username, hash)
	if err != nil {
		return err
	}

	return nil
}

func CreateRepo(repo string, username string) error {
	if isValid, err := isRepoNameValid(repo); !isValid {
		return err
	}

	b, err := repoAlreadyExist(username, repo)
	if err != nil {
		return err
	}
	if b {
		return errors.New("repo with the same name already exist")
	}

	statement, err := db.Prepare("insert into repo(userID,name,creation,public) VALUES(?,?,strftime('%s', 'now'),0)")
	if err != nil {
		return err
	}

	id, err := getUserID(username)
	if err != nil {
		return err
	}

	_, err = statement.Exec(id, repo)
	if err != nil {
		return err
	}

	return nil
}

func DeleteRepo(repo string, username string, signature string) error {
	err := VerifySignature(username, signature)
	if err != nil {
		return err
	}
	id, err := getUserID(username)
	if err != nil {
		return err
	}
	statement, err := db.Exec("delete FROM repo WHERE name=? AND userID=?", repo, id)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New(strconv.Itoa(int(rows)) + " deleted instead of only one")
	}
	return nil
}

func GetUsername(signature string) (string, bool) {
	username, exist := users[signature]
	return username, exist
}

func VerifySignature(username string, sig string) error {
	uname, b := GetUsername(sig)
	if !b || username != uname {
		return errors.New("invalid signature")
	}
	return nil
}

type Repo struct {
	RepoID   int
	UserID   int
	Username string
	Name     string
	Date     int
	IsPublic bool
}

func GetRepo(reponame string, username string) (Repo, error) {
	rows, err := db.Query("SELECT a.repoID, a.userID, a.name, a.creation, a.public FROM repo a INNER JOIN user b ON a.userID=b.userID WHERE UPPER(a.name) LIKE UPPER(?) AND UPPER(b.name) LIKE UPPER(?)", username, reponame)
	if err != nil {
		return Repo{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var ID int
		var uID int
		var name string
		var date int
		var public bool
		err = rows.Scan(&ID, &uID, &name, &date, &public)
		if err != nil {
			return Repo{}, err
		}
		return Repo{ID, uID, username, name, date, public}, nil
	}
	return Repo{}, errors.New("No repository called " + reponame + " by user " + username)
}

func GetRepoFromUser(username string, onlyPublic bool) ([]Repo, error) {
	var rows *sql.Rows
	var err error
	if onlyPublic {
		rows, err = db.Query("SELECT a.repoID, a.userID, a.name, a.creation, a.public FROM repo a INNER JOIN user b ON a.userID=b.userID WHERE UPPER(b.name) LIKE UPPER(?) AND a.public=1", username)
	} else {
		rows, err = db.Query("SELECT a.repoID, a.userID, a.name, a.creation, a.public FROM repo a INNER JOIN user b ON a.userID=b.userID WHERE UPPER(b.name) LIKE UPPER(?)", username)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []Repo
	for rows.Next() {
		var ID int
		var uID int
		var name string
		var date int
		var public bool
		err = rows.Scan(&ID, &uID, &name, &date, &public)
		if err != nil {
			return nil, err
		}
		repos = append(repos, Repo{ID, uID, username, name, date, public})
	}
	return repos, nil
}

func GetPublicRepo() ([]Repo, error) {
	rows, err := db.Query("SELECT b.name, a.repoID, a.userID, a.name, a.creation, a.public FROM repo a INNER JOIN user b ON a.userID=b.userID WHERE a.public=1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []Repo
	for rows.Next() {
		var username string
		var ID int
		var uID int
		var name string
		var date int
		var public bool
		err = rows.Scan(&username, &ID, &uID, &name, &date, &public)
		if err != nil {
			return nil, err
		}
		repos = append(repos, Repo{ID, uID, username, name, date, public})
	}
	return repos, nil
}

func IsRepoPublic(repo string, username string) (bool, error) {
	rows, err := db.Query("SELECT a.public FROM repo a INNER JOIN user b ON a.userID=b.userID WHERE UPPER(a.name) LIKE UPPER(?) AND UPPER(b.name) LIKE UPPER(?)", repo, username)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var public bool
		err = rows.Scan(&public)
		if err != nil {
			return false, err
		}
		return public, nil
	}
	return false, errors.New("No repository called " + repo + " by user " + username)
}

func TogglePublic(repo string, username string) error {
	statement, err := db.Prepare("UPDATE repo SET public=? WHERE UPPER(name) LIKE UPPER(?) AND userID=?")
	if err != nil {
		return err
	}
	id, err := getUserID(username)
	if err != nil {
		return err
	}
	b, err := IsRepoPublic(repo, username)
	if err != nil {
		return err
	}
	i := 1
	if b {
		i = 0
	}
	_, err = statement.Exec(i, repo, id)
	if err != nil {
		return err
	}
	return nil
}
