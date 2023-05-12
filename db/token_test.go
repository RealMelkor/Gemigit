package db

import (
	"testing"
	"time"
	"gemigit/test"
)

func TestCreateToken(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	_, err := user.CreateToken()
	test.IsNil(t, err)
}

func TestRenewToken(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	test.IsNotNil(t, user.RenewToken(-1), "token id should be invalid")

	_, err := user.CreateToken()
	test.IsNil(t, err)

	tokens, err := user.GetTokens()
	test.IsNil(t, err)

	test.IsEqual(t, len(tokens), 1)

	test.IsNil(t, user.RenewToken(tokens[0].ID))
}

func TestDeleteToken(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	test.IsNotNil(t, user.DeleteToken(-1), "token id should be invalid")

	_, err := user.CreateToken()
	test.IsNil(t, err)

	tokens, err := user.GetTokens()
	test.IsNil(t, err)
	test.IsEqual(t, len(tokens), 1)

	test.IsNil(t, user.DeleteToken(tokens[0].ID))
}

func TestGetTokens(t *testing.T) {
	
	initDB(t)

	user, _ := createUserAndSession(t)

	user.GetTokens()

	tokens, err := user.GetTokens()
	test.IsNil(t, err)
	test.IsEqual(t, len(tokens), 0)

	first, err := user.CreateToken()
	test.IsNil(t, err)

	second, err := user.CreateToken()
	test.IsNil(t, err)

	tokens, err = user.GetTokens()
	test.IsNil(t, err)
	test.IsEqual(t, len(tokens), 2)

	test.IsEqual(t, tokens[0].Hint, first[0:4])
	test.IsEqual(t, tokens[1].Hint, second[0:4])

	test.IsNil(t, user.DeleteToken(tokens[0].ID))

	tokens, err = user.GetTokens()
	test.IsNil(t, err)
	test.IsEqual(t, len(tokens), 1)
	test.IsEqual(t, tokens[0].Hint, second[0:4])
}

func TestCanUsePassword(t *testing.T) {
	
	initDB(t)

	_, err := CanUsePassword("invalid", "invalid", "invalid")
	test.IsNotNil(t, err, "should return user not found")

	user, signature := createUserAndSession(t)
	repo := test.FuncName(t)

	_, err = CanUsePassword("invalid", user.Name, user.Name)
	test.IsNotNil(t, err, "should return repository not found")

	test.IsNil(t, user.CreateRepo(repo, signature))

	b, err := CanUsePassword(repo, user.Name, user.Name)
	test.IsNil(t, err)
	test.IsEqual(t, b, true)

	test.IsNil(t, user.ToggleSecure())

	b, err = CanUsePassword(repo, user.Name, user.Name)
	test.IsNil(t, err)
	test.IsEqual(t, b, false)

	test.IsNil(t, user.ToggleSecure())

	b, err = CanUsePassword(repo, user.Name, user.Name)
	test.IsNil(t, err)
	test.IsEqual(t, b, true)
}

func TestTokenAuth(t *testing.T) {

	initDB(t)

	user, _ := createUserAndSession(t)

	test.IsNotNil(t, TokenAuth(user.Name, "invalid"),
			"token should be invalid")

	token, err := user.CreateToken()
	test.IsNil(t, err)
	test.IsNil(t, TokenAuth(user.Name, token))

	tokens, err := user.GetTokens()
	test.IsNil(t, err)
	test.IsEqual(t, len(tokens), 1)

	_, err = db.Exec("UPDATE token SET expiration = ? WHERE tokenID = ?",
			time.Now().Unix() - 1, tokens[0].ID)
	test.IsNil(t, err)
	test.IsNotNil(t, TokenAuth(user.Name, token),
			"token should be expired")
}
