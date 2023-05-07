package db

import (
	"errors"
)

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
			     "VALUES(?, ?, \"\", " + unixTime + ")",
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
	if len(desc) >= descriptionMaxLength {
		return errors.New("description too long")
	}
	res, err := db.Exec("UPDATE groups SET description = ? " +
			  "WHERE groupID = ?", desc, group)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows < 1 {
		return errors.New("group not found")
	}
	return err
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
	} else {
		return Member{}, errors.New("invalid group")
	}
	return m, nil
}

func (user User) GetMembers(group string) ([]Member, error) {
	query := "SELECT b.Name, b.UserID FROM member a " +
		 "INNER JOIN user b ON a.userID=b.userID " +
		 "INNER JOIN groups c ON a.groupID=c.groupID " +
		 "WHERE c.name = ?"
	rows, err := db.Query(query, group)
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
	if len(members) == 0 {
		return nil, errors.New("invalid group")
	}
	return members, nil
}

func (user User) IsInGroup(group string) (bool, error) {
	id, err := GetGroupID(group)
	if err != nil {
		return false, err
	}
	return user.IsInGroupID(id)
}

func (user User) GetGroups() ([]Group, error) {
	query := "SELECT a.groupID, a.name, a.description FROM groups a " +
		 "INNER JOIN member b ON a.groupID = b.groupID " +
		 "WHERE b.userID = ?"
	rows, err := db.Query(query, user.ID)
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
