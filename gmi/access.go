package gmi

import (
	"gemigit/db"
	"gemigit/csrf"

	"github.com/pitr/gig"
)

func accessRedirect(c gig.Context) error {
	return c.NoContent(gig.StatusRedirectTemporary,
		"/account/" + csrf.Token(c.CertHash()) + "/repo/" +
		c.Param("repo") + "/access")
}

func privilegeUpdate(privilege int, first bool) int {
	if first {
		return (privilege + 1)%3
	}
	if privilege == 0 {
		return 2
	}
	return privilege - 1
}

func privilegeToString(privilege int) string {
	switch (privilege) {
	case 0:
		return "none"
	case 1:
		return "read"
	case 2:
		return "read and write"
	}
	return "Invalid value"
}

func accessFirstOption(privilege int) string {
	switch (privilege) {
	case 0:
                return "Grant read access"
	case 1:
                return "Grant write access"
	default:
                return "Revoke read and write access"
	}
}

func accessSecondOption(privilege int) string {
	switch (privilege) {
	case 0:
                return "Grant read and write access"
	case 1:
                return "Revoke read access"
	default:
                return "Revoke write access"
	}
}

func changeGroupAccess(user db.User, repository string,
		       name string, first bool) error {
	repo, err := user.GetRepo(repository)
	if err != nil {
		return err
	}
	groupID, err := db.GetGroupID(name)
	if err != nil {
		return err
	}
	privilege, err := db.GetGroupAccess(repo, groupID)
	if err != nil {
		return err
	}
	privilege = privilegeUpdate(privilege, first)
	err = user.SetGroupAccess(repo, groupID, privilege)
	return err
}

func groupAccessOption(c gig.Context, first bool) error {
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest, "Invalid group")
        }
	err := changeGroupAccess(user, c.Param("repo"),
				 c.Param("group"), first)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accessRedirect(c)
}

func GroupAccessFirstOption(c gig.Context) error {
	return groupAccessOption(c, true)
}

func GroupAccessSecondOption(c gig.Context) error {
	return groupAccessOption(c, false)
}

func changeUserAccess(owner db.User, repository string,
		      name string, first bool) error {
	repo, err := owner.GetRepo(repository)
	if err != nil {
		return err
	}
	user, err := db.GetPublicUser(name)
	if err != nil {
		return err
	}
	privilege, err := db.GetUserAccess(repo, user)
	if err != nil {
		return err
	}
	privilege = privilegeUpdate(privilege, first)
	err = owner.SetUserAccess(repo, user.ID, privilege)
	return err
}

func userAccessOption(c gig.Context, first bool) error {
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest, "Invalid username")
        }
	err := changeUserAccess(user, c.Param("repo"), c.Param("user"), first)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accessRedirect(c)
}

func UserAccessFirstOption(c gig.Context) error {
	return userAccessOption(c, true)
}

func UserAccessSecondOption(c gig.Context) error {
	return userAccessOption(c, false)
}

func addAcess(c gig.Context, param string) (string, db.User, db.Repo, error) {
	query, err := c.QueryString()
	if err != nil {
		return "", db.User{}, db.Repo{},
		       c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if query == "" {
		return "", db.User{}, db.Repo{},
		       c.NoContent(gig.StatusInput, param)
	}
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return query, user, db.Repo{},
		       c.NoContent(gig.StatusBadRequest, "Invalid entry")
        }

	repo, err := user.GetRepo(c.Param("repo"))
	if err != nil {
                return query, user, db.Repo{},
		       c.NoContent(gig.StatusBadRequest,
				   "Repository not found")
	}
	return query, user, repo, nil
}

func AddUserAccess(c gig.Context) error {
	query, user, repo, err := addAcess(c, "User")
	if err != nil {
		return err
	}
	err = user.AddUserAccess(repo, query)
	if err != nil {
                return c.NoContent(gig.StatusBadRequest,
                                   "Invalid user")
	}
	return accessRedirect(c)
}

func AddGroupAccess(c gig.Context) error {
	query, user, repo, err := addAcess(c, "Group")
	if err != nil {
		return err
	}
	err = user.AddGroupAccess(repo, query)
	if err != nil {
                return c.NoContent(gig.StatusBadRequest,
                                   "Invalid user")
	}
	return accessRedirect(c)
}

func RemoveUserAccess(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest, "Invalid username")
        }
	userID, err := db.GetUserID(c.Param("user"))
	if err != nil {
                return c.NoContent(gig.StatusBadRequest,
                                   "User not found")
	}
	repo, err := user.GetRepo(c.Param("repo"))
	err = user.RemoveUserAccess(repo, userID)
	if err != nil {
                return c.NoContent(gig.StatusBadRequest,
                                   "User doesn't have access")
	}
	return accessRedirect(c)
}

func RemoveGroupAccess(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest, "Invalid username")
        }
	groupID, err := db.GetGroupID(c.Param("group"))
	if err != nil {
                return c.NoContent(gig.StatusBadRequest,
                                   "Group not found")
	}
	repo, err := user.GetRepo(c.Param("repo"))
	err = user.RemoveGroupAccess(repo, groupID)
	if err != nil {
                return c.NoContent(gig.StatusBadRequest,
                                   "Group doesn't have access")
	}
	return accessRedirect(c)
}
