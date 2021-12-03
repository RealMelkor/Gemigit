package auth

import (
	"errors"
	"gemigit/config"
	"gemigit/db"
	"time"
)

var userAttempts = make(map[string]int)
var clientAttempts = make(map[string]int)

func Decrease() {
	for {
		for k, v := range userAttempts {
			if v > 0 {
				userAttempts[k]--
			}
		}
		for k, v := range clientAttempts {
			if v > 0 {
				clientAttempts[k]--
			}
		}
		time.Sleep(30 * time.Second)
		db.DisconnectTimeout()
	}
}

func Connect(username string, password string, signature string, ip string) error {
	attempts, b := userAttempts[username]
	if b {
		if attempts < config.Cfg.Gemigit.MaxAttemptsForAccount {
			userAttempts[username]++
		} else {
			return errors.New("the account is locked, too many connections attempts")
		}
	} else {
		userAttempts[username] = 1
	}
	attempts, b = clientAttempts[ip]
	if b {
		if attempts < config.Cfg.Gemigit.MaxAttemptsForIP {
			clientAttempts[ip]++
		} else {
			return errors.New("too many connections attempts")
		}
	} else {
		clientAttempts[ip] = 1
	}
	b, err := db.Login(username, password, signature)
	if err != nil {
		return err
	}
	if !b {
		return errors.New("wrong username or password")
	}
	return nil
}
