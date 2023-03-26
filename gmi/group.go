package gmi

import (
	"gemigit/db"
	"github.com/pitr/gig"
)

func isGroupOwner(c gig.Context) (int, error) {
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return -1, c.NoContent(gig.StatusBadRequest,
					"Invalid username")
        }
	groupID, err := db.GetGroupID(c.Param("group"))
	if err != nil {
		return -1, c.NoContent(gig.StatusBadRequest, err.Error())
	}
	owner, err := user.IsInGroupID(groupID)
	if err != nil {
		return -1, c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if !owner {
                return -1, c.NoContent(gig.StatusBadRequest,
				       "Permission denied")
	}
	return groupID, nil
}

func SetGroupDesc(c gig.Context) error {
	query, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if query == "" {
		return c.NoContent(gig.StatusInput, "Description")
	}

	id, err := isGroupOwner(c)
	if err != nil {
		return err
	}

	err = db.SetGroupDescription(id, query)
	if err != nil {
		return err
	}
	return c.NoContent(gig.StatusRedirectTemporary,
			   "/account/groups/" + c.Param("group"))
}

func DeleteGroup(c gig.Context) error {
	name, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid input received")
	}
	if name == "" {
		return c.NoContent(gig.StatusInput,
				   "To confirm type the group name")
	}
	if name != c.Param("group") {
		return c.NoContent(gig.StatusRedirectTemporary,
				   "/account/groups/" + c.Param("group"))
	}
	id, err := isGroupOwner(c)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	err = db.DeleteGroup(id)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.NoContent(gig.StatusRedirectTemporary, "/account/groups")
}

func LeaveGroup(c gig.Context) (error) {
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

func RmFromGroup(c gig.Context) (error) {
	groupID, err := isGroupOwner(c)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	userID, err := db.GetUserID(c.Param("user"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	user, exist := db.GetUser(c.CertHash())
        if !exist {
                return c.NoContent(gig.StatusBadRequest,
                                   "Invalid username")
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
			   "/account/groups/" + c.Param("group"))
}

func AddToGroup(c gig.Context) (error) {
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
