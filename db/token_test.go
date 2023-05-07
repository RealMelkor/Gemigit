package db

import (
	"testing"
	"time"
)

func TestCreateToken(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	_, err := user.CreateToken()
	isNil(t, err)
}

func TestRenewToken(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	isNotNil(t, user.RenewToken(-1), "token id should be invalid")

	_, err := user.CreateToken()
	isNil(t, err)

	tokens, err := user.GetTokens()
	isNil(t, err)

	isEqual(t, len(tokens), 1)

	isNil(t, user.RenewToken(tokens[0].ID))
}

func TestDeleteToken(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	isNotNil(t, user.DeleteToken(-1), "token id should be invalid")

	_, err := user.CreateToken()
	isNil(t, err)

	tokens, err := user.GetTokens()
	isNil(t, err)
	isEqual(t, len(tokens), 1)

	isNil(t, user.DeleteToken(tokens[0].ID))
}

func TestGetTokens(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	user.GetTokens()

	tokens, err := user.GetTokens()
	isNil(t, err)
	isEqual(t, len(tokens), 0)

	first, err := user.CreateToken()
	isNil(t, err)

	second, err := user.CreateToken()
	isNil(t, err)

	tokens, err = user.GetTokens()
	isNil(t, err)
	isEqual(t, len(tokens), 2)

	isEqual(t, tokens[0].Hint, first[0:4])
	isEqual(t, tokens[1].Hint, second[0:4])

	isNil(t, user.DeleteToken(tokens[0].ID))

	tokens, err = user.GetTokens()
	isNil(t, err)
	isEqual(t, len(tokens), 1)
	isEqual(t, tokens[0].Hint, second[0:4])
}

func TestCanUsePassword(t *testing.T) {
	
	initDB(t)

	_, err := CanUsePassword("invalid", "invalid", "invalid")
	isNotNil(t, err, "should return user not found")

	user, signature := createUserAndSession(t)
	repo := funcName(t)

	_, err = CanUsePassword("invalid", user.Name, user.Name)
	isNotNil(t, err, "should return repository not found")

	isNil(t, user.CreateRepo(repo, signature))

	b, err := CanUsePassword(repo, user.Name, user.Name)
	isNil(t, err)
	isEqual(t, b, true)

	isNil(t, user.ToggleSecure())

	b, err = CanUsePassword(repo, user.Name, user.Name)
	isNil(t, err)
	isEqual(t, b, false)

	isNil(t, user.ToggleSecure())

	b, err = CanUsePassword(repo, user.Name, user.Name)
	isNil(t, err)
	isEqual(t, b, true)
}

func TestTokenAuth(t *testing.T) {

	initDB(t)

	user, _ := createUserAndSession(t)

	isNotNil(t, TokenAuth(user.Name, "invalid"), "token should be invalid")

	token, err := user.CreateToken()
	isNil(t, err)
	isNil(t, TokenAuth(user.Name, token))

	tokens, err := user.GetTokens()
	isNil(t, err)
	isEqual(t, len(tokens), 1)

	_, err = db.Exec("UPDATE token SET expiration = ? WHERE tokenID = ?",
			time.Now().Unix() - 1, tokens[0].ID)
	isNil(t, err)
	isNotNil(t, TokenAuth(user.Name, token), "token should be expired")
}
