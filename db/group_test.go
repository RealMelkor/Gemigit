package db

import (
	"testing"
)

const invalidGroupName = "mygroup$"

func TestCreateGroup(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	err := user.CreateGroup(group, signature + "a")
	isNotNil(t, err, "signature should be invalid")

	err = user.CreateGroup(invalidGroupName, signature)
	isNotNil(t, err, "group name should be invalid")

	isNil(t, user.CreateGroup(group, signature))

	err = user.CreateGroup(group, signature)
	isNotNil(t, err, "group name should already be taken")

}

func TestGetGroupID(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)
	group1 := funcName(t) + "a"
	group2 := funcName(t) + "b"
	
	isNil(t, user.CreateGroup(group1, signature))
	isNil(t, user.CreateGroup(group2, signature))

	id1, err := GetGroupID(group1)
	isNil(t, err)

	id2, err := GetGroupID(group2)
	isNil(t, err)

	id3, err := GetGroupID(group2)
	isNil(t, err)

	isNotEqual(t, id1, id2)
	isEqual(t, id2, id3)

	_, err = GetGroupID(group2 + "a")
	isNotNil(t, err, "group should be invalid")

}

func TestSetGroupDescription(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))

	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, SetGroupDescription(id, "description"))

	isNotNil(t, SetGroupDescription(-1, "description"),
			"group id should be invalid")

	isNotNil(t, SetGroupDescription(id, tooLongDescription),
			"description should be invalid")
}

func TestGetGroupDesc(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)
	group := funcName(t)
	description := group + "-description"

	isNil(t, user.CreateGroup(group, signature))

	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, SetGroupDescription(id, description))

	desc, err := GetGroupDesc(group)
	isNil(t, err)

	isEqual(t, desc, description)

	_, err = GetGroupDesc(group + "a")
	isNotNil(t, err, "group should be invalid")

}

func TestGetGroupOwner(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))

	owner, err := GetGroupOwner(group)
	isNil(t, err)

	isEqual(t, owner.Name, user.Name)

	_, err = GetGroupOwner(group + "a")
	isNotNil(t, err, "group should be invalid")

}

func TestAddUserToGroup(t *testing.T) {

	initDB(t)

	member1, _, _ := createUserAndSession(t)
	member2, _, _ := createUserAndSession(t)
	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))

	isNil(t, AddUserToGroup(group, member1.Name))
	isNil(t, AddUserToGroup(group, member2.Name))
	isNotNil(t, AddUserToGroup(group, member2.Name),
			"user should already be in the group")
	isNotNil(t, AddUserToGroup(group, member2.Name + "a"),
			"user should be invalid")
	isNotNil(t, AddUserToGroup(group + "a", member2.Name),
			"group should be invalid")

}

func TestDeleteGroup(t *testing.T) {

	initDB(t)

	member1, _, _ := createUserAndSession(t)
	member2, _, _ := createUserAndSession(t)
	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, AddUserToGroup(group, member1.Name))
	isNil(t, DeleteGroup(id))
	isNotNil(t, DeleteGroup(id), "group should be invalid")
	isNotNil(t, AddUserToGroup(group, member2.Name),
			"group should be invalid")

}

func TestDeleteMember(t *testing.T) {

	initDB(t)

	member1, _, _ := createUserAndSession(t)
	member2, _, _ := createUserAndSession(t)
	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, AddUserToGroup(group, member1.Name))
	isNil(t, AddUserToGroup(group, member2.Name))

	isNil(t, DeleteMember(member1.ID, id))
	isNotNil(t, DeleteMember(member1.ID, id),
		"user should already be deleted")

	isNil(t, DeleteMember(member2.ID, id))
	isNotNil(t, DeleteMember(member1.ID, id),
		"user should already be deleted")

	isNotNil(t, DeleteMember(member1.ID, -1),
		"group should be invalid")

	isNotNil(t, DeleteMember(-1, id),
		"user should be invalid")

}

func TestIsInGroupID(t *testing.T) {

	initDB(t)

	member1, _, _ := createUserAndSession(t)
	member2, _, _ := createUserAndSession(t)
	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, AddUserToGroup(group, member1.Name))

	b, err := member1.IsInGroupID(id)
	isNil(t, err)
	isEqual(t, b, false)

	b, err = member2.IsInGroupID(id)
	isNotNil(t, err, "should not be a member")

	b, err = user.IsInGroupID(id)
	isNil(t, err)
	isEqual(t, b, true)

}

func TestIsInGroup(t *testing.T) {

	initDB(t)

	member1, _, _ := createUserAndSession(t)
	member2, _, _ := createUserAndSession(t)
	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))

	isNil(t, AddUserToGroup(group, member1.Name))

	b, err := member1.IsInGroup(group)
	isNil(t, err)
	isEqual(t, b, false)

	b, err = member2.IsInGroup(group)
	isNotNil(t, err, "should not be a member")

	b, err = user.IsInGroup(group)
	isNil(t, err)
	isEqual(t, b, true)

	b, err = member2.IsInGroup(group + "a")
	isNotNil(t, err, "group should be invalid")
}

func TestGetGroups(t *testing.T) {

	initDB(t)

	member1, _, _ := createUserAndSession(t)
	member2, _, _ := createUserAndSession(t)
	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))
	isNil(t, user.CreateGroup(group + "2", signature))

	isNil(t, AddUserToGroup(group, member1.Name))

	groups, err := user.GetGroups()
	isNil(t, err)
	isEqual(t, len(groups), 2)

	groups, err = member1.GetGroups()
	isNil(t, err)
	isEqual(t, len(groups), 1)

	groups, err = member2.GetGroups()
	isNil(t, err)
	isEqual(t, len(groups), 0)

}

func TestGetMembers(t *testing.T) {

	initDB(t)

	member1, _, _ := createUserAndSession(t)
	user, _, signature := createUserAndSession(t)
	group := funcName(t)

	isNil(t, user.CreateGroup(group, signature))
	isNil(t, user.CreateGroup(group + "2", signature))

	members, err := user.GetMembers(group)
	isNil(t, err)
	isEqual(t, len(members), 1)

	isNil(t, AddUserToGroup(group, member1.Name))

	members, err = user.GetMembers(group)
	isNil(t, err)
	isEqual(t, len(members), 2)

	members, err = user.GetMembers(group + "a")
	isNotNil(t, err, "group should be invalid")
}
