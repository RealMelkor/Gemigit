package main

import (
	//"gemigit/config"
	"gemigit/db"

	"github.com/pitr/gig"
)

func leaveGroup(c gig.Context) (error) {
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest,
                                   "Invalid username")
        }
	groupID, err := db.GetGroupID(c.Param("group"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	owner, err := user.IsInGroupID(groupID)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if owner {
                return c.NoContent(gig.StatusBadRequest,
				   "You cannot leave your own group")
	}
	err = db.DeleteMember(user.ID, groupID)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.NoContent(gig.StatusRedirectTemporary, "/account/groups")
}

func rmFromGroup(c gig.Context) (error) {
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest,
                                   "Invalid username")
        }
	group := c.Param("group")
	groupID, err := db.GetGroupID(group)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	owner, err := user.IsInGroupID(groupID)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if !owner {
                return c.NoContent(gig.StatusBadRequest, "Permission denied")
	}
	userID, err := db.GetUserID(c.Param("user"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if userID == user.ID {
		return c.NoContent(gig.StatusBadRequest,
			"You cannot remove yourself from your own group")
	}
	err = db.DeleteMember(userID, groupID)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.NoContent(gig.StatusRedirectTemporary,
			   "/account/groups/" + group)
}

func addToGroup(c gig.Context) (error) {
	query, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if query == "" {
		return c.NoContent(gig.StatusInput, "Username")
	}

	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest,
                                   "Invalid username")
        }
	
	group := c.Param("group")
	owner, err := user.IsInGroup(group)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if !owner {
                return c.NoContent(gig.StatusBadRequest, "Permission denied")
	}

	if err = db.AddUserToGroup(group, query); err != nil {
                return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.NoContent(gig.StatusRedirectTemporary,
			   "/account/groups/" + group)
}
