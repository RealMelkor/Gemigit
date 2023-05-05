package db

import (
	"testing"
	"time"
)

func TestCreateToken(t *testing.T) {
	
	initDB(t)

	user, _, _ := createUserAndSession(t)

	_, err := user.CreateToken()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenewToken(t *testing.T) {
	
	initDB(t)

	user, _, _ := createUserAndSession(t)

	if err := user.RenewToken(-1); err == nil {
		t.Fatal("token id should be invalid")
	}

	_, err := user.CreateToken()
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := user.GetTokens()
	if err != nil {
		t.Fatal(err)
	}

	if len(tokens) != 1 {
		t.Fatal("the user should only have one token")
	}

	if err := user.RenewToken(tokens[0].ID); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteToken(t *testing.T) {
	
	initDB(t)

	user, _, _ := createUserAndSession(t)

	if err := user.DeleteToken(-1); err == nil {
		t.Fatal("token id should be invalid")
	}

	_, err := user.CreateToken()
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := user.GetTokens()
	if err != nil {
		t.Fatal(err)
	}

	if len(tokens) != 1 {
		t.Fatal("the user should only have one token")
	}

	if err := user.DeleteToken(tokens[0].ID); err != nil {
		t.Fatal(err)
	}
}

func TestGetTokens(t *testing.T) {
	
	initDB(t)

	user, _, _ := createUserAndSession(t)

	user.GetTokens()

	tokens, err := user.GetTokens()
	if err != nil {
		t.Fatal(err)
	}

	if len(tokens) != 0 {
		t.Fatal("the user should not have any token")
	}

	first, err := user.CreateToken()
	if err != nil {
		t.Fatal(err)
	}

	second, err := user.CreateToken()
	if err != nil {
		t.Fatal(err)
	}

	tokens, err = user.GetTokens()
	if err != nil {
		t.Fatal(err)
	}

	if len(tokens) != 2 {
		t.Fatal("the user should have 2 tokens")
	}

	if tokens[0].Hint != first[0:4] || tokens[1].Hint != second[0:4] {
		t.Fatal("invalid hint value")
	}

	if err = user.DeleteToken(tokens[0].ID); err != nil {
		t.Fatal(err)
	}

	tokens, err = user.GetTokens()
	if err != nil {
		t.Fatal(err)
	}

	if len(tokens) != 1 {
		t.Fatal("the user should have 2 tokens")
	}

	if tokens[0].Hint != second[0:4] {
		t.Fatal("invalid hint value")
	}
}

func TestCanUsePassword(t *testing.T) {
	
	initDB(t)

	_, err := CanUsePassword("invalid", "invalid", "invalid")
	if err == nil {
		t.Fatal("should return user not found")
	}

	user, _, signature := createUserAndSession(t)
	repo := funcName(t)

	_, err = CanUsePassword("invalid", user.Name, user.Name)
	if err == nil {
		t.Fatal("should return repository not found")
	}

	if err := user.CreateRepo(repo, signature); err != nil {
		t.Fatal(err)
	}

	b, err := CanUsePassword(repo, user.Name, user.Name)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("should be true by default")
	}

	if err := user.ToggleSecure(); err != nil {
		t.Fatal(err)
	}

	b, err = CanUsePassword(repo, user.Name, user.Name)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatal("should be false when secure mode is enabled")
	}

	if err := user.ToggleSecure(); err != nil {
		t.Fatal(err)
	}

	b, err = CanUsePassword(repo, user.Name, user.Name)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("should be true when secure mode is disabled")
	}
}

func TestTokenAuth(t *testing.T) {

	initDB(t)

	user, _, _ := createUserAndSession(t)

	if err := TokenAuth(user.Name, "invalid"); err == nil {
		t.Fatal("token should be invalid")
	}

	token, err := user.CreateToken()
	if err != nil {
		t.Fatal(err)
	}

	if err := TokenAuth(user.Name, token); err != nil {
		t.Fatal(err)
	}

	tokens, err := user.GetTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatal("the user should have one token")
	}

	_, err = db.Exec("UPDATE token SET expiration = ? WHERE tokenID = ?",
			time.Now().Unix() - 1, tokens[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := TokenAuth(user.Name, token); err == nil {
		t.Fatal("token should be expired")
	}
}
