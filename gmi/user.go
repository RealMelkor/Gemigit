package gmi

import (
        "gemigit/db"
        "gemigit/repo"
        "github.com/pitr/gig"
)

func ChangeDesc(c gig.Context) error {
	newdesc, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid input received")
	}
	if newdesc == "" {
		return c.NoContent(gig.StatusInput,
				   "New account description")
	}
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid username")
	}
	if err := user.ChangeDescription(newdesc, c.CertHash());
	   err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accountRedirect(c, "")
}

func AddRepo(c gig.Context) error {
	name, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if name == "" {
		return c.NoContent(gig.StatusInput, "Repository name")
	}
	user, b := db.GetUser(c.CertHash())
	if !b {
		return c.NoContent(gig.StatusBadRequest,
				   "Cannot find username")
	}
	if err := user.CreateRepo(name, c.CertHash());
	   err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   err.Error())
	}
	if err := repo.InitRepo(name, user.Name); err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   err.Error())
	}
	return accountRedirect(c, "repo/" + name)
}

func AddGroup(c gig.Context) error {
	name, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if name == "" {
		return c.NoContent(gig.StatusInput, "Group name")
	}
	user, b := db.GetUser(c.CertHash())
	if !b {
		return c.NoContent(gig.StatusBadRequest,
				   "Cannot find username")
	}
	if err := user.CreateGroup(name, c.CertHash());
	   err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   err.Error())
	}
	return accountRedirect(c, "groups/" + name)
}

func ChangePassword(c gig.Context) error {
	passwd, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid input received")
	}
	if passwd == "" {
		return c.NoContent(gig.StatusSensitiveInput, "New password")
	}
	user, b := db.GetUser(c.CertHash())
	if !b {
		return c.NoContent(gig.StatusBadRequest,
				   "Cannot find username")
	}
	err = user.ChangePassword(passwd, c.CertHash())
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accountRedirect(c, "")
}

func Disconnect(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	if err := user.Disconnect(c.CertHash()); err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.NoContent(gig.StatusRedirectTemporary, "/")
}

func DisconnectAll(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	if err := user.DisconnectAll(c.CertHash()); err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accountRedirect(c, "")
}
