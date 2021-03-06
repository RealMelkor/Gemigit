package db

import (
	"database/sql"
	"errors"
	"gemigit/config"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
}

var users = make(map[string]User)

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

func (user *User) repoAlreadyExist(repo string) (bool, error) {
	rows, err := db.Query("SELECT * FROM repo WHERE UPPER(name) LIKE UPPER(?)" +
			      " AND UPPER(userID) LIKE UPPER(?)", repo, user.ID)
	if err != nil {
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func DisconnectTimeout() {
	for k, v := range users {
		if time.Now().Unix()-v.Connection.Unix() >
		   int64(config.Cfg.Gemigit.AuthTimeout) {
			delete(users, k)
		}
	}
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
		"name" TEXT UNIQUE,
		"password" TEXT,
		"description" TEXT DEFAULT "",
		"creation" INTEGER		
	  );`

	createRepoTable := `CREATE TABLE repo (
		"repoID" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"userID" integer,
		"name" TEXT,
		"description" TEXT DEFAULT "",
		"creation" INTEGER,
		"public" INTEGER DEFAULT 0	
	);`

	_, err := db.Exec(createUserTable)
	if err != nil {
		return err
	}
	log.Println("Users table created")
	_, err = db.Exec(createRepoTable)
	if err != nil {
		return err
	}
	log.Println("Repositories table created")
	return nil
}

func CheckAuth(username string, password string) (bool, error) {
	rows, err := db.Query("select name, password from user" +
			      " WHERE UPPER(name) LIKE UPPER(?)", username)
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
	rows, err := db.Query("select userID, name, description, creation, password" +
			      " from user WHERE UPPER(name) LIKE UPPER(?)", username)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var dID int
		var dName string
		var dDescription string
		var dCreation int
		var dPassword string
		err = rows.Scan(&dID, &dName, &dDescription, &dCreation, &dPassword)
		if err != nil {
			return false, err
		}
		if checkPassword(password, dPassword) {
			users[signature] = User{ID: dID, Name: dName,
						Description: dDescription,
						Registration: dCreation,
						Connection: time.Now(),
						Signature: signature}
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

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = db.Exec("insert into user(name,password,creation) " +
			 "VALUES(?,?,strftime('%s', 'now'));", username, hash)
	if err != nil {
		return err
	}

	return nil
}

func (user User) CreateRepo(repo string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}

	if isValid, err := isRepoNameValid(repo); !isValid {
		return err
	}

	b, err := user.repoAlreadyExist(repo)
	if err != nil {
		return err
	}
	if b {
		return errors.New("repo with the same name already exist")
	}

	_, err = db.Exec("insert into repo(userID,name,creation,public,description)" +
			 " VALUES(?,?,strftime('%s', 'now'),0,\"\")", user.ID, repo)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUser(username string) error {
	statement, err := db.Exec("delete FROM repo " +
				  "WHERE userID in " +
				  "( SELECT userID from user " +
				  "where UPPER(name) LIKE UPPER(?))", username)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	statement, err = db.Exec("delete FROM user WHERE name=?", username)
	if err != nil {
		return err
	}
	rows, err = statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("user " + username + " not found")
	}
	return nil
}

func (user User) DeleteRepo(repo string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	statement, err := db.Exec("delete FROM repo WHERE name=? AND userID=?", repo, user.ID)
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

func GetUser(signature string) (User, bool) {
	user, b := users[signature]
	return user, b
}

func GetPublicUser(name string) (User, error) {
	rows, err := db.Query("select userID, name, description, creation" +
			      " from user WHERE UPPER(name) LIKE UPPER(?)", name)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var dID int
		var dName string
		var dDescription string
		var dCreation int
		err = rows.Scan(&dID, &dName, &dDescription, &dCreation)
		if err != nil {
			return User{}, err
		}
		return User{ID: dID,
			    Name: dName,
			    Description: dDescription,
			    Registration: dCreation}, nil
	}
	return User{}, errors.New(name + ", user not found")
}

func (user User) GetRepo(reponame string) (Repo, error) {
	rows, err := db.Query("SELECT repoID, userID, name," + 
			      " creation, public, description" +
			      " FROM repo WHERE UPPER(name) LIKE UPPER(?)" +
			      " AND userID=?", reponame, user.ID)
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
		var description string
		err = rows.Scan(&ID, &uID, &name, &date, &public, &description)
		if err != nil {
			return Repo{}, err
		}
		return Repo{ID, uID, user.Name, name, date, public, description}, nil
	}
	return Repo{}, errors.New("No repository called " + 
				  reponame + " by user " + user.Name)
}

func (user User) GetRepos(onlyPublic bool) ([]Repo, error) {
	var rows *sql.Rows
	var err error
	query := "SELECT repoID, userID, name, creation, public, description " + 
		 "FROM repo WHERE userID=?"
	if onlyPublic {
		query += " AND public=1"
	}
	rows, err = db.Query(query, user.ID)
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
		var description string
		err = rows.Scan(&ID, &uID, &name, &date, &public, &description)
		if err != nil {
			return nil, err
		}
		repos = append(repos, Repo{ID, uID, user.Name, name,
					   date, public, description})
	}
	return repos, nil
}

func GetPublicRepo() ([]Repo, error) {
	rows, err := db.Query("SELECT b.name, a.repoID, a.userID, a.name, " +
			      "a.creation, a.public, a.description " +
			      "FROM repo a INNER JOIN user b " +
			      "ON a.userID=b.userID WHERE a.public=1")
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
		var description string
		err = rows.Scan(&username, &ID, &uID, &name,
				&date, &public, &description)
		if err != nil {
			return nil, err
		}
		repos = append(repos, Repo{ID, uID, username,
					   name, date, public, description})
	}
	return repos, nil
}

func IsRepoPublic(repo string, username string) (bool, error) {
	rows, err := db.Query("SELECT a.public FROM repo a " +
			      "INNER JOIN user b ON a.userID=b.userID " +
			      "WHERE UPPER(a.name) LIKE UPPER(?) " +
			      "AND UPPER(b.name) LIKE UPPER(?)", repo, username)
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

func (user User) TogglePublic(repo string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	b, err := IsRepoPublic(repo, user.Name)
	if err != nil {
		return err
	}
	i := 1
	if b {
		i = 0
	}
	_, err = db.Exec("UPDATE repo SET public=? " +
			 "WHERE UPPER(name) LIKE UPPER(?) " + 
			 "AND userID=?", i, repo, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (user *User) VerifySignature(signature string) error {
	if user.Signature != signature {
		return errors.New("wrong signature")
	}
	if users[signature].ID != user.ID {
		return errors.New("signature doesn't match the user")
	}
	return nil
}

func ChangePassword(username string, password string) error {
	b, err := isPasswordValid(password)
	if err != nil {
		return err
	}
	if !b {
		return errors.New("invalid password")
	}
	hPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	statement, err := db.Exec("UPDATE user SET password=? " + 
				  "WHERE UPPER(name) LIKE UPPER(?)",
				  hPassword, username)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("unknown user " + username)
	}
	return nil
}

func (user User) ChangePassword(password string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	return ChangePassword(user.Name, password)
}

func (user User) ChangeDescription(description string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	statement, err := db.Exec("UPDATE user SET description=? " +
				  "WHERE UPPER(name) LIKE UPPER(?)",
				  description, user.Name)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("no description changed")
	}
	u, b := users[signature]
	if !b {
		return errors.New("invalid signature detected")
	}
	u.Description = description
	users[signature] = u
	return nil
}

func (user User) Disconnect(signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	delete(users, signature)
	return nil
}

func (user User) ChangeRepoName(name string, newname string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	b, err := isRepoNameValid(newname)
	if err != nil {
		return err
	}
	if !b {
		return errors.New("invalid name")
	}
	statement, err := db.Exec("UPDATE repo SET name=? " +
				  "WHERE UPPER(name) LIKE UPPER(?) AND userID=?",
				  newname, name, user.ID)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("failed to change the repository name")
	}
	return nil
}

func (user User) ChangeRepoDesc(name string, newdesc string) error {
	statement, err := db.Exec("UPDATE repo SET description=?" +
				  " WHERE UPPER(name) LIKE UPPER(?) " +
				  " AND userID=?", newdesc, name, user.ID)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("failed to change the repository description")
	}
	return nil
}

func GetRepoDesc(name string, username string) (string, error) {
	rows, err := db.Query("SELECT a.description FROM repo a " +
			      "INNER JOIN user b ON a.userID=b.userID " +
			      "WHERE UPPER(a.name) LIKE UPPER(?) " +
			      "AND UPPER(b.name) LIKE UPPER(?)", name, username)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if rows.Next() {
		var description string
		err = rows.Scan(&description)
		if err != nil {
			return "", err
		}
		return description, nil
	}
	return "", errors.New("No repository called " + name + " by user " + username)
}

func (user *User) UpdateDescription() error {
	rows, err := db.Query("select description from user WHERE userID=?", user.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		var dDescription string
		err = rows.Scan(&dDescription)
		if err != nil {
			return err
		}
		user.Description = dDescription
	}
	users[user.Signature] = *user
	return nil
}
