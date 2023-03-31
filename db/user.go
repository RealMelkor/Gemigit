package db

import (
	"errors"
	"time"
	"gemigit/config"
)

var users = make(map[string]User)

func (user *User) VerifySignature(signature string) error {
	if user.Signature != signature {
		return errors.New("wrong signature")
	}
	if users[signature].ID != user.ID {
		return errors.New("signature doesn't match the user")
	}
	return nil
}

func userAlreadyExist(username string) error {
	rows, err := db.Query(
		"SELECT * FROM user WHERE UPPER(name) LIKE UPPER(?)",
		username)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("This username is already taken")
	}
	return nil
}

func GetPublicUser(name string) (User, error) {
	rows, err := db.Query("SELECT userID, name, description, creation " +
			      "FROM user WHERE UPPER(name) LIKE UPPER(?)",
			      name)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return User{}, errors.New(name + ", user not found")
	}
	var u = User{}
	err = rows.Scan(&u.ID, &u.Name,
			&u.Description,
			&u.Registration)
	if err != nil {
		return User{}, err
	}
	return u, nil
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

func (user *User) UpdateDescription() error {
	rows, err := db.Query("SELECT description FROM user WHERE userID=?",
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

func (user *User) SetUserSecret(secret string) error {
	_, err := db.Exec("UPDATE user SET secret = ? " +
			  "WHERE userID = ?", secret, user.ID)
	user.Secret = secret
	users[user.Signature] = *user
	return err
}

func DeleteUser(username string) error {
	statement, err := db.Exec("DELETE FROM repo " +
				  "WHERE userID in " +
				  "(SELECT userID FROM user " +
				  "where UPPER(name) LIKE UPPER(?))",
				  username)
	if err != nil {
		return err
	}
	rows, err := statement.RowsAffected()
	if err != nil {
		return err
	}
	statement, err = db.Exec("DELETE FROM user WHERE name=?", username)
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

func GetUser(signature string) (User, bool) {
	user, b := users[signature]
	// should update description
	if b {
		return user, b
	}
	rows, err := db.Query(`SELECT a.userID, name, description, a.creation,
				a.secret, a.securegit
				FROM user a INNER JOIN certificate b ON
				a.userID = b.userID WHERE b.hash = ?`,
				signature)
	if err != nil {
		return User{}, false
	}
	defer rows.Close()
	if !rows.Next() {
		return User{}, false
	}
	err = rows.Scan(&user.ID, &user.Name, &user.Description,
			&user.Registration, &user.Secret, &user.SecureGit)
	if err != nil {
		return User{}, false
	}
	user.Signature = signature
	users[signature] = user
	return user, true
}

func FetchUser(username string, signature string) (User, error) {
	query := `SELECT userID, name, description, creation, secret, securegit
			FROM user WHERE UPPER(name) LIKE UPPER(?)`
	rows, err := db.Query(query, username)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	next := rows.Next()
	if !next {
		return User{}, errors.New("User not found")
	}
	var u = User{}
	err = rows.Scan(&u.ID,
			&u.Name,
			&u.Description,
			&u.Registration,
			&u.Secret,
			&u.SecureGit)
	if err != nil {
		return User{}, err
	}
	u.Connection = time.Now()
	u.Signature = signature
	return u, nil
}

func CheckAuth(username string, password string) (error) {
	rows, err := db.Query("SELECT name, password FROM user " +
			      "WHERE UPPER(name) LIKE UPPER(?)",
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

func Register(username string, password string) error {

	if !config.Cfg.Ldap.Enabled {
		if err := isPasswordValid(password); err != nil {
			return err
		}
	}

	if err := isUsernameValid(username); err != nil {
		return err
	}

	if err := userAlreadyExist(username); err != nil {
		return err
	}

	if !config.Cfg.Ldap.Enabled {
		hash, err := hashPassword(password)
		if err != nil {
			return err
		}

		_, err = db.Exec("INSERT INTO user(name,password,creation) " +
				 "VALUES(?, ?, " + unixTime + ");",
				 username, hash)
		return err
	}
	_, err := db.Exec("INSERT INTO user(name,creation) " +
			  "VALUES(?, " + unixTime + ")", username)
	return err
}
