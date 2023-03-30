package db

import (
	"errors"
	"strconv"
)

func (user *User) repoAlreadyExist(repo string) error {
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

	_, err = db.Exec("INSERT INTO repo " +
			 "(userID, name, creation, public, description) " +
			 "VALUES(?, ?, " + unixTime + ", 0, \"\")",
			 user.ID, repo)
	if err != nil {
		return err
	}

	return nil
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

func (user User) DeleteRepo(repo string, signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	statement, err := db.Exec("DELETE FROM repo WHERE name=? AND userID=?",
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
	return err
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
	query := "SELECT repoID, userID, name, " +
		 "creation, public, description " +
		 "FROM repo WHERE userID=?"
	if onlyPublic {
		query += " AND public=1"
	}
	rows, err := db.Query(query, user.ID)
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
