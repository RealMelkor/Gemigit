package gmi

import (
	"log"
	"strconv"

	"gemigit/db"
	"gemigit/csrf"

	"github.com/pitr/gig"
)

func tokenRedirect(c gig.Context) error {
	return c.NoContent(gig.StatusRedirectTemporary,
		"/account/" + csrf.Token(c.CertHash()) + "/token")
}

func CreateToken(c gig.Context, readOnly bool) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	token, err := user.CreateToken(readOnly)
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}
	data := struct {
		Token string
	}{
		Token: token,
	}
	return execT(c, "token_new.gmi", data)
}

func CreateWriteToken(c gig.Context) error {
	return CreateToken(c, false)
}

func CreateReadToken(c gig.Context) error {
	return CreateToken(c, true)
}

func ListTokens(c gig.Context) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	tokens, err := user.GetTokens()
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}

	data := struct {
		Tokens []db.Token
		Secure bool
	}{
		Tokens: tokens,
		Secure: user.SecureGit,
	}
	return execT(c, "token.gmi", data)
}

func ToggleTokenAuth(c gig.Context) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	err := user.ToggleSecure()
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}

	return tokenRedirect(c)
}

func RenewToken(c gig.Context) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	id, err := strconv.Atoi(c.Param("token"))
	if err != nil || user.RenewToken(id) != nil {
		return c.NoContent(gig.StatusBadRequest, "Invalid token")
	}

	return tokenRedirect(c)
}

func DeleteToken(c gig.Context) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	id, err := strconv.Atoi(c.Param("token"))
	if err != nil || user.DeleteToken(id) != nil {
		return c.NoContent(gig.StatusBadRequest, "Invalid token")
	}

	return tokenRedirect(c)
}
