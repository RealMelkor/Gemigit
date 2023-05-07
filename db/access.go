package db

import (
	"errors"
	"log"
)

func GetUserGroupAccess(repoID int, userID int) (int, error) {
	rows, err := db.Query("SELECT a.privilege FROM access a " +
			      "INNER JOIN member m ON a.groupID = m.groupID " +
			      "WHERE a.repoID = ? AND m.userID = ?",
			      repoID, userID)
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

func GetUserAccess(repoID int, userID int) (int, error) {
	rows, err := db.Query("SELECT privilege FROM access " +
			     "WHERE repoID = ? AND userID = ? ",
			     repoID, userID)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, errors.New("The user is not a contributor")
	}
	var privilege int
	err = rows.Scan(&privilege)
	if err != nil {
		return -1, err
	}
	return privilege, nil
}

func GetAccess(repoID int, userID int) (int, error) {
	privilege, err := GetUserGroupAccess(repoID, userID)
	if err != nil {
		return -1, err
	}
	if privilege == 2 {
		return 2, nil
	}
	p, err := GetUserAccess(repoID, userID)
	if err != nil {
		return -1, err
	}
	if p < privilege {
		return privilege, nil
	}
	return p, nil
}

func SetUserAccess(repoID int, userID int, privilege int) (error) {
	_, err := db.Exec("UPDATE access SET privilege = ? " +
			  "WHERE repoID = ? AND userID = ?",
			  privilege, repoID, userID)
	return err
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

func HasReadAccessTo(userID int) ([]Repo, error) {
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
			      userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	repos := []Repo{}
	for rows.Next() {
		var r Repo
		err = rows.Scan(&r.RepoID, &r.UserID, &r.Name, &r.Date,
				&r.IsPublic, &r.Description, &r.Username)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, nil
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
	return err
}

func (u *User) RemoveGroupAccess(repo Repo, groupID int) error {

	if repo.UserID != u.ID {
		return errors.New(
			"Only the repository owner can revoke access")
	}

	statement, err := db.Exec("DELETE FROM access " +
				  "WHERE groupID = ? AND repoID = ?",
				  groupID, repo.RepoID)
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
			      repo.RepoID, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("The user already has access")
	}

	_, err = db.Exec("INSERT INTO access (repoID, userID, privilege) " +
			 "VALUES(?, ?, 1)", repo.RepoID, userID)
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
			      repo.RepoID, groupID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("The group already has access")
	}

	_, err = db.Exec("INSERT INTO access (repoID, groupID, privilege) " +
			 "VALUES(?, ?, 1)", repo.RepoID, groupID)
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
				  userID, repo.RepoID)
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
