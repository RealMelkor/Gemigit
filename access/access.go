package access

import (
	"errors"
	"fmt"
	"gemigit/config"
	"gemigit/db"

	ldap "github.com/go-ldap/ldap/v3"
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
func Login(name string, password string) (error) {
	if name == "" || password == "" {
		return errors.New("empty field")
	}
	if config.Cfg.Ldap.Enabled {
		err := conn.Bind(fmt.Sprintf("%s=%s,%s",
				 config.Cfg.Ldap.Attribute,
				 ldap.EscapeFilter(name),
				 config.Cfg.Ldap.Binding),
				 password)
		return err
	}
	err := db.CheckAuth(name, password)
	if err != nil {
		return err
	}
	return nil
}

