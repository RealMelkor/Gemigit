package db

import (
	"errors"
	"log"
)

const accessNone = 0
const accessRead = 1
const accessReadWrite = 2
const accessDefault = 1
const accessMax = 2

func GetUserGroupAccess(repo Repo, user User) (int, error) {
	rows, err := db.Query("SELECT a.privilege FROM access a " +
			      "INNER JOIN member m ON a.groupID = m.groupID " +
			      "WHERE a.repoID = ? AND m.userID = ?",
			      repo.ID, user.ID)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
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
	return privilege, nil
}

func GetUserAccess(repo Repo, user User) (int, error) {
	if user.ID == repo.UserID {
		return accessMax, nil
	}
	rows, err := db.Query("SELECT privilege FROM access " +
			     "WHERE repoID = ? AND userID = ? ",
			     repo.ID, user.ID)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return accessNone, nil
	}
	var privilege int
	err = rows.Scan(&privilege)
	if err != nil {
		return -1, err
	}
	return privilege, nil
}

func GetAccess(user User, repo Repo) (int, error) {
	privilege, err := GetUserGroupAccess(repo, user)
	if err != nil {
		return -1, err
	}
	if privilege == accessMax {
		return accessMax, nil
	}
	p, err := GetUserAccess(repo, user)
	if err != nil {
		return -1, err
	}
	if p < privilege {
		return privilege, nil
	}
	return p, nil
}

func (u User) SetUserAccess(repo Repo, userID int, privilege int) (error) {
	if repo.UserID != u.ID {
		return errors.New(
			"only the repository owner can manage access")
	}
	if u.ID == userID {
		return errors.New(
			"the repository owner access cannot be changed")
	}
	_, err := db.Exec("UPDATE access SET privilege = ? " +
			  "WHERE repoID = ? AND userID = ?",
			  privilege, repo.ID, userID)
	return err
}

func GetRepoUserAccess(repoID int) ([]Access, error) {
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
		err = rows.Scan(&a.RepoID, &a.GroupID, &a.Name, &a.Privilege)
		if err != nil {
			return nil, err
		}
		access = append(access, a)
	}
	return access, nil
}

func (u User) HasReadAccessTo() ([]Repo, error) {
	rows, err := db.Query("SELECT r.repoID, r.userID, r.name, " +
			      "r.creation, r.public, r.description, u.Name " +
			      "FROM access a " +
			      "INNER JOIN member m ON a.groupID = m.groupID " +
			      "INNER JOIN repo r ON a.repoID = r.repoID " +
			      "INNER JOIN user u ON r.userID = u.userID " +
			      "WHERE m.userID = ? AND a.privilege > 0 " +
			      "AND ? <> r.userID " +
			      "UNION " +
			      "SELECT r.repoID, r.userID, r.name, " +
			      "r.creation, r.public, r.description, u.Name " +
			      "FROM access a " +
			      "INNER JOIN repo r ON a.repoID = r.repoID " +
			      "INNER JOIN user u ON r.userID = u.userID " +
			      "WHERE a.userID = ? AND a.privilege > 0 " +
			      "AND ? <> r.userID",
			      u.ID, u.ID, u.ID, u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	repos := []Repo{}
	for rows.Next() {
		var r Repo
		err = rows.Scan(&r.ID, &r.UserID, &r.Name, &r.Date,
				&r.IsPublic, &r.Description, &r.Username)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, nil
}

func GetGroupAccess(repo Repo, groupID int) (int, error) {
	rows, err := db.Query("SELECT privilege FROM access " +
			      "WHERE repoID = ? AND groupID = ? ",
			      repo.ID, groupID)
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

func (u User) SetGroupAccess(repo Repo, groupID int, privilege int) (error) {
	if repo.UserID != u.ID {
		return errors.New(
			"only the repository owner can manage access")
	}
	_, err := db.Exec("UPDATE access SET privilege = ? " +
			  "WHERE repoID = ? AND groupID = ?",
			  privilege, repo.ID, groupID)
	return err
}

func (u User) RemoveGroupAccess(repo Repo, groupID int) error {

	if repo.UserID != u.ID {
		return errors.New(
			"Only the repository owner can revoke access")
	}

	statement, err := db.Exec("DELETE FROM access " +
				  "WHERE groupID = ? AND repoID = ?",
				  groupID, repo.ID)
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

func (u *User) AddUserAccess(repo Repo, user string) error {

	if repo.UserID != u.ID {
		return errors.New(
			"Only the repository owner can add members")
	}

	userID, err := GetUserID(user)
	if err != nil {
		return err
	}
	
	if userID == u.ID {
		return errors.New(
			"The repository owner already has maximum privilege")
	}

	rows, err := db.Query("SELECT privilege FROM access " +
			      "WHERE repoID = ? AND userID = ? ",
			      repo.ID, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("The user already has access")
	}

	_, err = db.Exec("INSERT INTO access (repoID, userID, privilege) " +
			 "VALUES(?, ?, 1)", repo.ID, userID)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (u *User) AddGroupAccess(repo Repo, group string) error {

	if repo.UserID != u.ID {
		return errors.New(
			"Only the repository owner can add groups")
	}

	groupID, err := GetGroupID(group)
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT privilege FROM access " +
			      "WHERE repoID = ? AND groupID = ? ",
			      repo.ID, groupID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("The group already has access")
	}

	_, err = db.Exec("INSERT INTO access (repoID, groupID, privilege) " +
			 "VALUES(?, ?, 1)", repo.ID, groupID)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (u *User) RemoveUserAccess(repo Repo, userID int) error {
	if u.ID != repo.UserID {
		return errors.New(
			"Only the repository owner can revoke access")
	}
	statement, err := db.Exec("DELETE FROM access " +
				  "WHERE userID = ? AND repoID = ?",
				  userID, repo.ID)
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
