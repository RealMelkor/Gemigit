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

var users = make(map[string]User)

func userAlreadyExist(username string) (bool, error) {
	rows, err := db.Query(
		"select * from user WHERE UPPER(name) LIKE UPPER(?)",
		username)
	if err != nil {
		return true, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func (user *User) repoAlreadyExist(repo string) (error) {
	rows, err := db.Query(
		"SELECT * FROM repo WHERE UPPER(name) LIKE UPPER(?)" +
		" AND UPPER(userID) LIKE UPPER(?)",
		repo, user.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New(
			"Repository with the same name already exist")
	}
	return nil
}

func groupAlreadyExist(group string) (error) {
	rows, err := db.Query(
		"SELECT * FROM groups WHERE UPPER(name) LIKE UPPER(?)",
		group)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New(
			"A group with the same name already exist")
	}
	return nil
}

func DisconnectTimeout() {
	if config.Cfg.Users.Session == 0 {
		return
	}
	for k, v := range users {
		if time.Now().Unix()-v.Connection.Unix() >
		   int64(config.Cfg.Users.Session) {
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
	userTable := `CREATE TABLE user (
		"userID" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT UNIQUE NOT NULL,
		"password" TEXT,
		"description" TEXT DEFAULT "",
		"creation" INTEGER NOT NULL
	);`

	groupTable := `CREATE TABLE groups (
		"groupID" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"owner" INTEGER NOT NULL,
		"name" TEXT UNIQUE NOT NULL,
		"description" TEXT DEFAULT "",
		"creation" INTEGER NOT NULL
	);`

	memberTable := `CREATE TABLE member (
		"groupID" INTEGER NOT NULL,
		"userID" INTEGER NOT NULL
	);`

	accessTable := `CREATE TABLE access (
		"repoID" INTEGER NOT NULL,
		"groupID" INTEGER,
		"userID" INTEGER,
		"privilege" INTEGER NOT NULL
	);`

	repoTable := `CREATE TABLE repo (
		"repoID" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"userID" INTEGER NOT NULL,
		"name" TEXT NOT NULL,
		"description" TEXT DEFAULT "",
		"creation" INTEGER NOT NULL,
		"public" INTEGER DEFAULT 0
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
	log.Println("Member table created")

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

	return nil
}

func CheckAuth(username string, password string) (error) {
	rows, err := db.Query("select name, password from user" +
			      " WHERE UPPER(name) LIKE UPPER(?)",
			      username)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		var dPassword string
		var dName string
		err = rows.Scan(&dName, &dPassword)
		if err != nil {
			return err
		}
		if checkPassword(password, dPassword) {
			return nil
		}
	}
	return errors.New("invalid credential")
}

func FetchUser(username string, signature string) (User, error) {
	query := "select userID, name, description, creation " +
		 "from user WHERE UPPER(name) LIKE UPPER(?)"
	rows, err := db.Query(query, username)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	next := rows.Next()
	if !next {
		return User{}, errors.New("")
	}
	var u = User{}
	var dPassword string
	if config.Cfg.Ldap.Enabled {
		err = rows.Scan(&u.ID,
				&u.Name,
				&u.Description,
				&u.Registration)
	} else {
		err = rows.Scan(&u.ID,
				&u.Name,
				&u.Description,
				&u.Registration,
				&dPassword)
	}
	if err != nil {
		return User{}, err
	}
	u.Connection = time.Now()
	u.Signature = signature
	return u, nil
}

func Register(username string, password string) error {

	if !config.Cfg.Ldap.Enabled {
		if err := isPasswordValid(password); err != nil {
			return err
		}
	}

	if err := isNameValid(username); err != nil {
		return err
	}

	if exist, err := userAlreadyExist(username); exist || err != nil {
		if err != nil {
			return err
		}
		return errors.New("this name is already taken")
	}

	if !config.Cfg.Ldap.Enabled {
		hash, err := hashPassword(password)
		if err != nil {
			return err
		}

		_, err = db.Exec("insert into user(name,password,creation) " +
				 "VALUES(?,?,strftime('%s', 'now'));",
				 username, hash)
		if err != nil {
			return err
		}
		return nil
	}
	_, err := db.Exec("insert into user(name,creation) " +
			  "VALUES(?,strftime('%s', 'now'));", username)
	if err != nil {
		return err
	}
	return nil
}

func (user User) CreateRepo(repo string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}

	if err := isRepoNameValid(repo); err != nil {
		return err
	}

	err := user.repoAlreadyExist(repo)
	if err != nil {
		return err
	}

	_, err = db.Exec("insert into repo " +
			 "(userID, name, creation, public, description) " +
			 "VALUES(?, ?, strftime('%s', 'now'), 0, \"\")",
			 user.ID, repo)
	if err != nil {
		return err
	}

	return nil
}

func (user User) CreateGroup(group string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}

	if err := isGroupNameValid(group); err != nil {
		return err
	}

	err := groupAlreadyExist(group)
	if err != nil {
		return err
	}

	rows, err := db.Exec("INSERT INTO groups " +
			     "(owner, name, description, creation) " +
			     "VALUES(?, ?, \"\", strftime('%s', 'now'))",
			     user.ID, group)
	if err != nil {
		return err
	}
	
	groupID, err := rows.LastInsertId()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO member (groupID, userID) " +
			 "VALUES(?, ?)", groupID, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func GetUserID(name string) (int, error) {
	query := "SELECT userID FROM user WHERE UPPER(?) = UPPER(name);"
	rows, err := db.Query(query, name)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, errors.New("User not found")
	}
	var id int
	err = rows.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func GetRepoID(repo string, userID int) (int, error) {
	rows, err := db.Query("SELECT repoID FROM repo " +
			      "WHERE UPPER(?) = UPPER(name) AND userID = ?",
			      repo, userID)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, errors.New("Repository not found")
	}
	var id int
	err = rows.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func GetGroupDesc(name string) (string, error) {
	query := "SELECT description FROM groups WHERE UPPER(?) = UPPER(name);"
	rows, err := db.Query(query, name)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if !rows.Next() {
		return "", errors.New("Group not found")
	}
	var desc string
	err = rows.Scan(&desc)
	if err != nil {
		return "", err
	}
	return desc, nil
}

func GetGroupID(name string) (int, error) {
	query := "SELECT groupID FROM groups WHERE UPPER(?) = UPPER(name);"
	rows, err := db.Query(query, name)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, errors.New("Group not found")
	}
	var id int
	err = rows.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func IsInGroup(userID int, groupID int) (error) {
	query := "SELECT * FROM member WHERE userID=? AND groupID=?"
	rows, err := db.Query(query, userID, groupID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("The user is already in the group")
	}
	return nil
}

func AddUserToGroup(group string, user string) error {
	id, err := GetGroupID(group)
	if err != nil {
		return err
	}
	userID, err := GetUserID(user)
	if err != nil {
		return err
	}
	if err = IsInGroup(userID, id); err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO member (groupID, userID) " +
			  "VALUES(?, ?)", id, userID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteMember(user int, group int) error {
	statement, err := db.Exec("DELETE FROM member " +
				  "WHERE userID = ? AND groupID = ?",
				  user, group)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("The user is not a member of the group")
	}
	return nil
}

func SetGroupDescription(group int, desc string) error {
	_, err := db.Exec("UPDATE groups SET description = ? " +
			  "WHERE groupID = ?", desc, group)
	if err != nil {
		return err
	}
	return nil
}

func DeleteGroup(group int) error {
	statement, err := db.Exec("DELETE FROM groups " +
				  "WHERE groupID = ?", group)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("There's no such group")
	}
	statement, err = db.Exec("DELETE FROM member " +
				 "WHERE groupID = ?", group)
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(username string) error {
	statement, err := db.Exec("delete FROM repo " +
				  "WHERE userID in " +
				  "(SELECT userID from user " +
				  "where UPPER(name) LIKE UPPER(?))",
				  username)
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
	statement, err := db.Exec("delete FROM repo WHERE name=? AND userID=?",
				  repo, user.ID)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New(strconv.Itoa(int(rows)) +
				  " deleted instead of only one")
	}
	return nil
}

func SetUser(signature string, user User) {
	users[signature] = user
}

func GetUser(signature string) (User, bool) {
	user, b := users[signature]
	return user, b
}

func GetPublicUser(name string) (User, error) {
	rows, err := db.Query("select userID, name, description, creation " +
			      "from user WHERE UPPER(name) LIKE UPPER(?)",
			      name)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var u = User{}
		err = rows.Scan(&u.ID, &u.Name,
				&u.Description,
				&u.Registration)
		if err != nil {
			return User{}, err
		}
		return u, nil
	}
	return User{}, errors.New(name + ", user not found")
}

func AddUserAccess(owner int, repoID int, user string) error {
	userID, err := GetUserID(user)
	if err != nil {
		return err
	}
	if userID == owner {
		return errors.New(
			"The repository owner already has maximum privilege")
	}

	rows, err := db.Query("SELECT privilege FROM access " +
			      "WHERE repoID = ? AND userID = ? ",
			      repoID, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("The user already has access")
	}

	_, err = db.Exec("INSERT INTO access (repoID, userID, privilege) " +
			 "VALUES(?, ?, 1)", repoID, userID)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func AddGroupAccess(repoID int, group string) error {
	groupID, err := GetGroupID(group)
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT privilege FROM access " +
			      "WHERE repoID = ? AND groupID = ? ",
			      repoID, groupID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("The group already has access")
	}

	_, err = db.Exec("INSERT INTO access (repoID, groupID, privilege) " +
			 "VALUES(?, ?, 1)", repoID, groupID)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func RemoveUserAccess(repoID int, userID int) error {
	statement, err := db.Exec("DELETE FROM access " +
				  "WHERE userID = ? AND repoID = ?",
				  userID, repoID)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("The user is not a contributor")
	}
	return nil
}

func RemoveGroupAccess(repoID int, groupID int) error {
	statement, err := db.Exec("DELETE FROM access " +
				  "WHERE groupID = ? AND repoID = ?",
				  groupID, repoID)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("The group is not a contributor")
	}
	return nil

}

func GetGroupAccess(repoID int, groupID int) (int, error) {
	rows, err := db.Query("SELECT privilege FROM access " +
			      "WHERE repoID = ? AND groupID = ? ",
			      repoID, groupID)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, errors.New("The group is not a contributor")
	}
	var privilege int
	err = rows.Scan(&privilege)
	if err != nil {
		return -1, err
	}
	return privilege, nil
}

func SetGroupAccess(repoID int, groupID int, privilege int) (error) {
	_, err := db.Exec("UPDATE access SET privilege = ? " +
			  "WHERE repoID = ? AND groupID = ?",
			  privilege, repoID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func GetUserAccess(repoID int, userID int) (int, error) {
	rows, err := db.Query("SELECT a.privilege FROM access a " +
			      "INNER JOIN member m ON a.groupID = m.groupID " +
			      "WHERE a.repoID = 3 AND m.userID = 2",
			      repoID, userID)
	if err != nil {
		return -1, err
	}
	privilege := -1
	for rows.Next() {
		var p int
		err = rows.Scan(&p)
		if err != nil {
			return -1, err
		}
		if p > privilege {
			privilege = p
		}
	}
	rows.Close()
	if privilege == 2 {
		return 2, nil
	}

	rows, err = db.Query("SELECT privilege FROM access " +
			      "WHERE repoID = ? AND userID = ? ",
			      repoID, userID)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() && privilege == -1 {
		return -1, errors.New("The user is not a contributor")
	}
	var p int
	err = rows.Scan(&p)
	if err != nil {
		return -1, err
	}
	if privilege < p {
		privilege = p
	}
	return privilege, nil
}

func SetUserAccess(repoID int, userID int, privilege int) (error) {
	_, err := db.Exec("UPDATE access SET privilege = ? " +
			  "WHERE repoID = ? AND userID = ?",
			  privilege, repoID, userID)
	if err != nil {
		return err
	}
	return nil
}

func GetRepoAccess(repoID int) ([]Access, error) {
	rows, err := db.Query("SELECT a.repoID, a.userID, b.name, " +
			      "a.privilege FROM access a " +
			      "INNER JOIN user b ON a.userID = b.userID " +
			      "WHERE a.repoID = ?", repoID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	var access []Access
	for rows.Next() {
		var a = Access{}
		err = rows.Scan(&a.RepoID, &a.UserID, &a.Name, &a.Privilege)
		if err != nil {
			return nil, err
		}
		access = append(access, a)
	}
	return access, nil
}

func GetRepoGroupAccess(repoID int) ([]Access, error) {
	rows, err := db.Query("SELECT a.repoID, a.groupID, b.name, " +
			      "a.privilege FROM access a " +
			      "INNER JOIN groups b ON a.groupID = b.groupID " +
			      "WHERE a.repoID = ?", repoID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	var access []Access
	for rows.Next() {
		var a = Access{}
		err = rows.Scan(&a.RepoID, &a.UserID, &a.Name, &a.Privilege)
		if err != nil {
			return nil, err
		}
		access = append(access, a)
	}
	return access, nil
}

func (user User) GetRepo(reponame string) (Repo, error) {
	rows, err := db.Query("SELECT repoID, userID, name, " + 
			      "creation, public, description " +
			      "FROM repo WHERE UPPER(name) LIKE UPPER(?) " +
			      "AND userID=?", reponame, user.ID)
	if err != nil {
		return Repo{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var r = Repo{}
		err = rows.Scan(&r.RepoID, &r.UserID, &r.Name,
				&r.Date, &r.IsPublic, &r.Description)
		if err != nil {
			return Repo{}, err
		}
		return r, nil
	}
	return Repo{}, errors.New("No repository called " + 
				  reponame + " by user " + user.Name)
}

func (user User) GetRepos(onlyPublic bool) ([]Repo, error) {
	var rows *sql.Rows
	var err error
	query := "SELECT repoID, userID, name, " +
		 "creation, public, description " + 
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
		var r = Repo{}
		err = rows.Scan(&r.RepoID, &r.UserID, &r.Name,
				&r.Date, &r.IsPublic, &r.Description)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, nil
}

func (user User) GetGroups() ([]Group, error) {
	var rows *sql.Rows
	var err error
	query := "SELECT a.groupID, a.name, a.description FROM groups a " +
		 "INNER JOIN member b ON a.groupID = b.groupID " +
		 "WHERE b.userID = ?"
	rows, err = db.Query(query, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Group
	for rows.Next() {
		var g = Group{}
		err = rows.Scan(&g.GroupID, &g.Name, &g.Description)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (user User) IsInGroup(group string) (bool, error) {
	id, err := GetGroupID(group)
	if err != nil {
		return false, err
	}
	return user.IsInGroupID(id)
}

func (user User) IsInGroupID(groupID int) (bool, error) {
	query := "SELECT owner FROM member a " +
		 "INNER JOIN groups b ON a.groupID = b.groupID " +
		 "WHERE a.userID = ? AND a.groupID = ? "
	rows, err := db.Query(query, user.ID, groupID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return false, errors.New("Group not found")
	}
	var owner int
	err = rows.Scan(&owner)
	if err != nil {
		return false, err
	}
	return owner == user.ID, nil
}

func GetGroupOwner(group string) (Member, error) {
	query := "SELECT c.name, a.userID FROM member a " +
		 "INNER JOIN groups b ON a.groupID = b.groupID " +
		 "INNER JOIN user c ON a.userID = c.userID " +
		 "WHERE a.userID = b.owner AND b.name = ? "
	rows, err := db.Query(query, group)
	if err != nil {
		return Member{}, err
	}
	defer rows.Close()
	var m = Member{}
	if rows.Next() {
		err = rows.Scan(&m.Name, &m.UserID)
		if err != nil {
			return Member{}, err
		}
	}
	return m, nil
}

func (user User) GetMembers(group string) ([]Member, error) {
	var rows *sql.Rows
	var err error
	query := "SELECT b.Name, b.UserID FROM member a " +
		 "INNER JOIN user b ON a.userID=b.userID " +
		 "INNER JOIN groups c ON a.groupID=c.groupID " +
		 "WHERE c.name = ?"
	rows, err = db.Query(query, group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []Member
	for rows.Next() {
		var m = Member{}
		err = rows.Scan(&m.Name, &m.UserID)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
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
		var r = Repo{}
		err = rows.Scan(&r.Username, &r.RepoID, &r.UserID, &r.Name,
				&r.Date, &r.IsPublic, &r.Description)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, nil
}

func IsRepoPublic(repo string, username string) (bool, error) {
	rows, err := db.Query("SELECT a.public FROM repo a " +
			      "INNER JOIN user b ON a.userID = b.userID " +
			      "WHERE UPPER(a.name) LIKE UPPER(?) " +
			      "AND UPPER(b.name) LIKE UPPER(?)",
			      repo, username)
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
	return false, errors.New("No repository called " + repo +
				 " by user " + username)
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
	err := isPasswordValid(password)
	if err != nil {
		return err
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

func (user User) ChangeDescription(desc string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	statement, err := db.Exec("UPDATE user SET description=? " +
				  "WHERE UPPER(name) LIKE UPPER(?)",
				  desc, user.Name)
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
	u.Description = desc
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

func (user User) ChangeRepoName(name string, newname string,
				signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	err := isRepoNameValid(newname)
	if err != nil {
		return err
	}
	statement, err := db.Exec("UPDATE repo SET name=? " +
				  "WHERE UPPER(name) LIKE UPPER(?) " +
				  "AND userID=?",
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
	statement, err := db.Exec("UPDATE repo SET description=? " +
				  "WHERE UPPER(name) LIKE UPPER(?) " +
				  "AND userID=?", newdesc, name, user.ID)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}
	return errors.New("failed to change the repository description")
}

func GetRepoDesc(name string, username string) (string, error) {
	rows, err := db.Query("SELECT a.description FROM repo a " +
			      "INNER JOIN user b ON a.userID=b.userID " +
			      "WHERE UPPER(a.name) LIKE UPPER(?) " +
			      "AND UPPER(b.name) LIKE UPPER(?)",
			      name, username)
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
	return "", errors.New("No repository called " + name +
			      " by user " + username)
}

func (user *User) UpdateDescription() error {
	rows, err := db.Query("select description from user WHERE userID=?",
			      user.ID)
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
