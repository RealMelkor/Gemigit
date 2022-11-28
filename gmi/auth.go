package gmi

import (
	"gemigit/db"
	"github.com/pitr/gig"
)

const textRegistrationSuccess =
          "# Your registration was completed successfully\n\n" +
          "=> /login Login now"

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
	if err = db.Register(c.Param("name"), password); err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.Gemini(textRegistrationSuccess)
}
