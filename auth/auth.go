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
var registrationAttempts = make(map[string]int)

func Decrease() {
	for {
		userAttempts = make(map[string]int)
		clientAttempts = make(map[string]int)
		registrationAttempts = make(map[string]int)
		time.Sleep(time.Duration(config.Cfg.Protection.Reset) *
			   time.Second)
	}
}

func try(attemps *map[string]int, key string, max int) bool {
	value, exist := (*attemps)[key]
	if exist {
		if value < max {
			(*attemps)[key]++
		} else {
			return true
		}
	} else {
		(*attemps)[key] = 1
	}
	return false
}

// Check if credential are valid then add client signature
// as a connected user
func Connect(username string, password string,
	     signature string, ip string) error {

	if try(&userAttempts, username, config.Cfg.Protection.Account) {
		return errors.New("the account is locked, " +
				  "too many connections attempts")
	}

	if try(&clientAttempts, ip, config.Cfg.Protection.Ip) {
		return errors.New("too many connections attempts")
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
	err = db.Register(username, "")
	if err != nil {
		return err
	}
	user, err = db.FetchUser(username, signature)
	if err != nil {
		return err
	}
	db.SetUser(signature, user)
	return nil
}

func Register(username string, password string, ip string) error {
	if try(&registrationAttempts, ip, config.Cfg.Protection.Registration) {
		return errors.New("too many registration attempts")
	}
	return db.Register(username, password)
}
