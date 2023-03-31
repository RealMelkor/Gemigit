package gmi

import (
	"gemigit/auth"
	"gemigit/config"
	"gemigit/db"

	"github.com/pitr/gig"
	"github.com/pquerna/otp/totp"

	"log"
	"bytes"
	"image/png"
)

var keys = make(map[string]string)

func CreateTOTP(c gig.Context) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer: config.Cfg.Title,
		AccountName: user.Name,
	})

	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}

	var buf bytes.Buffer
	img, err_ := key.Image(200, 200)
	if err_ != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}
	png.Encode(&buf, img)

	keys[c.CertHash()] = key.Secret()

	return c.Blob("image/png", buf.Bytes())
}

func ConfirmTOTP(c gig.Context) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	query, err := c.QueryString()
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}
	if query == "" {
		return c.NoContent(gig.StatusInput, "Code")
	}

	key, exist := keys[c.CertHash()]

	valid := false
	if exist {
		valid = totp.Validate(query, key)
	}
	if !valid {
		return c.NoContent(gig.StatusBadRequest, "Invalid code")
	}

	err = user.SetUserSecret(key)
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}

	return c.NoContent(gig.StatusRedirectTemporary, "/account")
}

func LoginOTP(c gig.Context) error {

	query, err := c.QueryString()
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}
	if query == "" {
		return c.NoContent(gig.StatusInput, "Code")
	}

	err = auth.LoginOTP(c.CertHash(), query)
	if err != nil && err.Error() == "wrong code" {
		return c.NoContent(gig.StatusInput, "Code")
	}
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.NoContent(gig.StatusRedirectTemporary, "/account")
}

func RemoveTOTP(c gig.Context) error {

	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}

	query, err := c.QueryString()
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}
	if query == "" {
		return c.NoContent(gig.StatusInput, "Code")
	}

	valid := totp.Validate(query, user.Secret)
	if !valid {
		return c.NoContent(gig.StatusInput, "Code")
	}

	err = user.SetUserSecret("")
	if err != nil {
		log.Println(err)
		return c.NoContent(gig.StatusBadRequest, "Unexpected error")
	}

	return c.NoContent(gig.StatusRedirectTemporary, "/account/otp")
}

func ShowOTP(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	data := struct {
		Secret bool
	}{
		Secret: user.Secret != "",
	}
	return execT(c, "otp.gmi", data)
}
