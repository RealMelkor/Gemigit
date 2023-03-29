package access

import (
	"errors"
	"fmt"
	"gemigit/config"
	"gemigit/db"

	ldap "github.com/go-ldap/ldap/v3"
)

const (
	None = 0
	Read = 1
	Write = 2
)

var conn *ldap.Conn

func Init() error {
	if !config.Cfg.Ldap.Enabled {
		return nil
	}
        l, err := ldap.DialURL(config.Cfg.Ldap.Url)
        if err != nil {
                return err
        }
        conn = l
        return nil
}

// return nil if credential are valid, an error if not
func Login(name string, password string, useToken bool) (error) {
	if name == "" || password == "" {
		return errors.New("empty field")
	}
	if useToken && db.TokenAuth(name, password) != nil{
		return nil
	}
	if config.Cfg.Ldap.Enabled {
		return conn.Bind(fmt.Sprintf("%s=%s,%s",
				 config.Cfg.Ldap.Attribute,
				 ldap.EscapeFilter(name),
				 config.Cfg.Ldap.Binding),
				 password)
	}
	return db.CheckAuth(name, password)
}

func hasAccess(repo string, author string, user string, access int) error {
	userID, err := db.GetUserID(user)
	if err != nil {
		return err
	}
	u, err := db.GetPublicUser(author)
	if err != nil {
		return err
	}
	r, err := u.GetRepo(repo)
	if err != nil {
		return err
	}
	if r.UserID == userID {
		return nil
	}
	privilege, err := db.GetAccess(r.RepoID, userID)
	if err != nil {
		return err
	}
	if privilege < access {
		return errors.New("Permission denied")
	}
	return nil
}

func HasWriteAccess(repo string, author string, user string) error {
	return hasAccess(repo, author, user, Write)
}

func HasReadAccess(repo string, author string, user string) error {
	return hasAccess(repo, author, user, Read)
}
