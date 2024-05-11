package gmi

import (
	"log"
	"strconv"

	"gemigit/db"

	"github.com/pitr/gig"
)

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

	return redirect(c, "token")
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

	return redirect(c, "token")
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

	return redirect(c, "token")
}
