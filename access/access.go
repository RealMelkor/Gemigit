package access

import (
	"errors"
	"fmt"
	"gemigit/config"
	//"gemigit/db"

	ldap "github.com/go-ldap/ldap/v3"
)

var conn *ldap.Conn

func Init() error {
        l, err := ldap.DialURL(config.Cfg.Ldap.Url)
        if err != nil {
                return err
        }
        conn = l
        return nil
}

func Login(name string, password string) (error) {
	if config.Cfg.Ldap.Enabled {
		if name == "" || password == "" {
			return errors.New("empty field")
		}
		err := conn.Bind(fmt.Sprintf("%s=%s,%s",
				 config.Cfg.Ldap.Attribute,
				 ldap.EscapeFilter(name),
				 config.Cfg.Ldap.Binding),
				 password)
		return err
	}
	return nil
	/*
	b, err := db.CheckAuth(name, password)
	if err != nil {
		return err
	}
	if !b {
		return errors.New("invalid credential")
	}
	return nil*/
}

