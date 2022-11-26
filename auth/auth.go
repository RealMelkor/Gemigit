package auth

import (
	"errors"
	"gemigit/config"
	"gemigit/db"
	"gemigit/access"
	"time"
)

var userAttempts = make(map[string]int)
var clientAttempts = make(map[string]int)

func Decrease() {
	for {
		userAttempts = make(map[string]int)
		clientAttempts = make(map[string]int)
		time.Sleep(time.Duration(config.Cfg.Protection.Reset) *
			   time.Second)
		db.DisconnectTimeout()
	}
}

// Check if credential are valid then add client signature
// as a connected user
func Connect(username string, password string,
	     signature string, ip string) error {
	attempts, b := userAttempts[username]
	if b {
		if attempts < config.Cfg.Protection.Account {
			userAttempts[username]++
		} else {
			return errors.New("the account is locked, " + 
					  "too many connections attempts")
		}
	} else {
		userAttempts[username] = 1
	}
	attempts, b = clientAttempts[ip]
	if b {
		if attempts < config.Cfg.Protection.Ip {
			clientAttempts[ip]++
		} else {
			return errors.New("too many connections attempts")
		}
	} else {
		clientAttempts[ip] = 1
	}
	err := access.Login(username, password)
	if err != nil {
		return err
	}

	user, err := db.FetchUser(username, signature)
	if err == nil {
		db.SetUser(signature, user)
		return nil
	}
	if !config.Cfg.Ldap.Enabled {
		return err
	}
	db.Register(username, "")
	user, err = db.FetchUser(username, signature)
	if err != nil {
		return err
	}
	db.SetUser(signature, user)
	return nil
}
