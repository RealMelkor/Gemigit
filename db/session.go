package db

import (
	"log"
)

func (user User) GetSessionsCount() (int, error) {
	row, err := db.Query("SELECT COUNT(userID) FROM certificate " +
				"WHERE userID = ?", user.ID)
	if err != nil {
		return -1, err
	}
	defer row.Close()
	count := 0
	if row.Next() {
		err = row.Scan(&count)
		if err != nil {
			return -1, err
		}
	}
	return count, nil
}

func (user User) Disconnect(signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	delete(users, signature)
	_, err := db.Exec("DELETE FROM certificate WHERE hash = ?", signature)
	if err != nil {
		return err
	}
	return nil
}

func (user User) DisconnectAll(signature string) error {
	if err := user.VerifySignature(signature); err != nil {
		return err
	}
	rows, err := db.Query("SELECT hash FROM certificate WHERE userID = ?",
				user.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		err = rows.Scan(&hash)
		if err != nil {
			return err
		}
		if hash != signature {
			delete(users, hash)
		}
	}
	_, err = db.Exec("DELETE FROM certificate WHERE userID = ? AND " +
				"hash <> ?", user.ID, signature)
	if err != nil {
		return err
	}
	return nil
}

func AddUserSession(signature string, user User) error {
	_, err := db.Exec("INSERT INTO certificate (userID, hash, creation) " +
			"VALUES (?, ?, " + unixTime + ")", user.ID, signature)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	user.Signature = signature
	users[signature] = user
	return nil
}
