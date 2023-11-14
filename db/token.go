package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
	"log"
	"errors"
)

func CanUsePassword(repo string, owner string, username string) (bool, error) {
	row, err := db.Query(`SELECT securegit FROM user WHERE
				UPPER(name) LIKE UPPER(?)`, username)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer row.Close()
	var secure int
	if (!row.Next()) {
		return false, errors.New("not found")
	}
	row.Scan(&secure)
	if secure != 0 {
		return false, nil
	}
	row.Close()

	row, err = db.Query(`SELECT b.securegit FROM user a
				INNER JOIN repo b ON a.userID = b.UserID WHERE
				UPPER(a.name) LIKE UPPER(?) AND
				UPPER(b.name) LIKE UPPER(?)`,
				owner, repo)
	if err != nil {
		log.Println(err)
		return false, err
	}
	if (!row.Next()) {
		return false, errors.New("not found")
	}
	defer row.Close()
	row.Scan(&secure)
	return secure == 0, nil
}

func (user User) CreateToken(readOnly bool) (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	token := base64.RawStdEncoding.EncodeToString(data)
	sum := sha256.Sum224(data)
	hash := base64.RawStdEncoding.EncodeToString(sum[:])
	_, err := db.Exec(`INSERT INTO
		token(userID, token, hint, expiration, readonly)
		VALUES(?, ?, ?, ?, ?);`,
		user.ID, hash, token[0:4], time.Now().Unix() + 3600 * 24 * 30,
		readOnly)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (user User) RenewToken(tokenID int) (error) {
	res, err := db.Exec(`UPDATE token SET expiration = ?
				WHERE tokenID = ? AND userID = ?`,
				time.Now().Unix() + 3600 * 24 * 30 + 1,
				tokenID, user.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows < 1 {
		return errors.New("invalid token id")
	}
	return nil
}

func (user User) DeleteToken(tokenID int) (error) {
	row, err := db.Exec(
		`DELETE FROM token WHERE tokenID = ? AND userID = ?`,
		tokenID, user.ID)
	if err != nil {
		return err
	}
	count, err := row.RowsAffected()
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New("invalid token")
	}
	return nil
}

func (user User) GetTokens() ([]Token, error) {
	rows, err := db.Query(`SELECT tokenID, expiration, hint, readonly
				FROM token WHERE userID = ?`, user.ID)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unexpected error")
	}
	defer rows.Close()

	tokens := []Token{}
	for rows.Next() {
		var r Token
		err = rows.Scan(&r.ID, &r.Expiration, &r.Hint, &r.ReadOnly)
		if err != nil {
			return nil, err
		}
		r.ExpirationFormat = time.Unix(r.Expiration, 0).UTC().
							Format(time.RFC1123)
		tokens = append(tokens, r)
	}
	return tokens, nil
}

func (user *User) ToggleSecure() error {
	user.SecureGit = !user.SecureGit

	_, err := db.Exec("UPDATE user SET securegit = ? " +
			  "WHERE userID = ?", user.SecureGit, user.ID)
	if err != nil {
		return err
	}

	users[user.Signature] = *user
	return nil
}

func TokenAuth(username string, token string, wantWrite bool) error {
	decoded, err := base64.RawStdEncoding.DecodeString(token)
	if err != nil {
		log.Println(err)
		return errors.New("invalid token")
	}
	sum := sha256.Sum224(decoded)
	hash := base64.RawStdEncoding.EncodeToString(sum[:])
	row, err := db.Query(`SELECT b.expiration, b.readonly FROM user a
				INNER JOIN token b ON a.userID = b.UserID WHERE
				UPPER(a.name) LIKE UPPER(?) AND
				UPPER(b.token) LIKE UPPER(?)`,
				username, hash)
	if err != nil {
		log.Println(err)
		return errors.New("unexpected error")
	}
	defer row.Close()
	if !row.Next() {
		return errors.New("invalid token")
	}
	var expiration int64
	var readonly bool
	row.Scan(&expiration, &readonly)
	if expiration <= time.Now().Unix() {
		return errors.New("token expired")
	}
	if wantWrite {
		if readonly {
			return errors.New("the token only has read access")
		}
	}
	return nil
}
