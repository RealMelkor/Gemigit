package gmi

import (
	"gemigit/auth"
	"github.com/pitr/gig"
)

func Register(c gig.Context) error {
	cert := c.Certificate()
	if cert == nil {
		return c.NoContent(gig.StatusClientCertificateRequired,
				   "Certificate required")
	}

	name, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid name received")
	}
	if name != "" {
		return c.NoContent(gig.StatusRedirectPermanent,
				   "/register/" + name)
	}

	return c.NoContent(gig.StatusInput, "Username")
}

func RegisterConfirm(c gig.Context) error {
	cert := c.Certificate()
	if cert == nil {
		return c.NoContent(gig.StatusClientCertificateRequired,
				   "Certificate required")
	}

	password, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid password received")
	}
	if password == "" {
		return c.NoContent(gig.StatusSensitiveInput, "Password")
	}
	if err = auth.Register(c.Param("name"), password, c.IP()); err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	data, err := execTemplate("register_success.gmi", nil)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.Gemini(data)
}

func Login(user, pass, sig string, c gig.Context) (string, error) {
	err := auth.Connect(user, pass, sig, c.IP())

	if err != nil && err.Error() == "token required" {
		return "/otp", nil
	}
	if err != nil {
		return "", err
	}
	return "/account", nil
}
