package db

import (
	"testing"
	"gemigit/test"
)

const invalidGroupName = "mygroup$"

func TestCreateGroup(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	err := user.CreateGroup(group, signature + "a")
	test.IsNotNil(t, err, "signature should be invalid")

	err = user.CreateGroup(invalidGroupName, signature)
	test.IsNotNil(t, err, "group name should be invalid")

	test.IsNil(t, user.CreateGroup(group, signature))

	err = user.CreateGroup(group, signature)
	test.IsNotNil(t, err, "group name should already be taken")

}

func TestGetGroupID(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	group1 := test.FuncName(t) + "a"
	group2 := test.FuncName(t) + "b"
	
	test.IsNil(t, user.CreateGroup(group1, signature))
	test.IsNil(t, user.CreateGroup(group2, signature))

	id1, err := GetGroupID(group1)
	test.IsNil(t, err)

	id2, err := GetGroupID(group2)
	test.IsNil(t, err)

	id3, err := GetGroupID(group2)
	test.IsNil(t, err)

	test.IsNotEqual(t, id1, id2)
	test.IsEqual(t, id2, id3)

	_, err = GetGroupID(group2 + "a")
	test.IsNotNil(t, err, "group should be invalid")

}

func TestSetGroupDescription(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))

	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, SetGroupDescription(id, "description"))

	test.IsNotNil(t, SetGroupDescription(-1, "description"),
			"group id should be invalid")

	test.IsNotNil(t, SetGroupDescription(id, tooLongDescription),
			"description should be invalid")
}

func TestGetGroupDesc(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	group := test.FuncName(t)
	description := group + "-description"

	test.IsNil(t, user.CreateGroup(group, signature))

	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, SetGroupDescription(id, description))

	desc, err := GetGroupDesc(group)
	test.IsNil(t, err)

	test.IsEqual(t, desc, description)

	_, err = GetGroupDesc(group + "a")
	test.IsNotNil(t, err, "group should be invalid")

}

func TestGetGroupOwner(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))

	owner, err := GetGroupOwner(group)
	test.IsNil(t, err)

	test.IsEqual(t, owner.Name, user.Name)

	_, err = GetGroupOwner(group + "a")
	test.IsNotNil(t, err, "group should be invalid")

}

func TestAddUserToGroup(t *testing.T) {

	initDB(t)

	member1, _ := createUserAndSession(t)
	member2, _ := createUserAndSession(t)
	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))

	test.IsNotNil(t, member1.AddUserToGroup(group, member1.Name),
			"only the group owner can add members")

	test.IsNil(t, user.AddUserToGroup(group, member1.Name))
	test.IsNil(t, user.AddUserToGroup(group, member2.Name))
	test.IsNotNil(t, user.AddUserToGroup(group, member2.Name),
			"user should already be in the group")
	test.IsNotNil(t, user.AddUserToGroup(group, member2.Name + "a"),
			"user should be invalid")
	test.IsNotNil(t, user.AddUserToGroup(group + "a", member2.Name),
			"group should be invalid")

}

func TestDeleteGroup(t *testing.T) {

	initDB(t)

	member1, _ := createUserAndSession(t)
	member2, _ := createUserAndSession(t)
	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.AddUserToGroup(group, member1.Name))
	test.IsNil(t, DeleteGroup(id))
	test.IsNotNil(t, DeleteGroup(id), "group should be invalid")
	test.IsNotNil(t, user.AddUserToGroup(group, member2.Name),
			"group should be invalid")

}

func TestDeleteMember(t *testing.T) {

	initDB(t)

	member1, _ := createUserAndSession(t)
	member2, _ := createUserAndSession(t)
	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.AddUserToGroup(group, member1.Name))
	test.IsNil(t, user.AddUserToGroup(group, member2.Name))

	test.IsNil(t, DeleteMember(member1.ID, id))
	test.IsNotNil(t, DeleteMember(member1.ID, id),
		"user should already be deleted")

	test.IsNil(t, DeleteMember(member2.ID, id))
	test.IsNotNil(t, DeleteMember(member1.ID, id),
		"user should already be deleted")

	test.IsNotNil(t, DeleteMember(member1.ID, -1),
		"group should be invalid")

	test.IsNotNil(t, DeleteMember(-1, id),
		"user should be invalid")

}

func TestIsInGroupID(t *testing.T) {

	initDB(t)

	member1, _ := createUserAndSession(t)
	member2, _ := createUserAndSession(t)
	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.AddUserToGroup(group, member1.Name))

	b, err := member1.IsInGroupID(id)
	test.IsNil(t, err)
	test.IsEqual(t, b, false)

	b, err = member2.IsInGroupID(id)
	test.IsNotNil(t, err, "should not be a member")

	b, err = user.IsInGroupID(id)
	test.IsNil(t, err)
	test.IsEqual(t, b, true)

}

func TestIsInGroup(t *testing.T) {

	initDB(t)

	member1, _ := createUserAndSession(t)
	member2, _ := createUserAndSession(t)
	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))

	test.IsNil(t, user.AddUserToGroup(group, member1.Name))

	b, err := member1.IsInGroup(group)
	test.IsNil(t, err)
	test.IsEqual(t, b, false)

	b, err = member2.IsInGroup(group)
	test.IsNotNil(t, err, "should not be a member")

	b, err = user.IsInGroup(group)
	test.IsNil(t, err)
	test.IsEqual(t, b, true)

	b, err = member2.IsInGroup(group + "a")
	test.IsNotNil(t, err, "group should be invalid")
}

func TestGetGroups(t *testing.T) {

	initDB(t)

	member1, _ := createUserAndSession(t)
	member2, _ := createUserAndSession(t)
	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))
	test.IsNil(t, user.CreateGroup(group + "2", signature))

	test.IsNil(t, user.AddUserToGroup(group, member1.Name))

	groups, err := user.GetGroups()
	test.IsNil(t, err)
	test.IsEqual(t, len(groups), 2)

	groups, err = member1.GetGroups()
	test.IsNil(t, err)
	test.IsEqual(t, len(groups), 1)

	groups, err = member2.GetGroups()
	test.IsNil(t, err)
	test.IsEqual(t, len(groups), 0)

}

func TestGetMembers(t *testing.T) {

	initDB(t)

	member1, _ := createUserAndSession(t)
	user, signature := createUserAndSession(t)
	group := test.FuncName(t)

	test.IsNil(t, user.CreateGroup(group, signature))
	test.IsNil(t, user.CreateGroup(group + "2", signature))

	members, err := user.GetMembers(group)
	test.IsNil(t, err)
	test.IsEqual(t, len(members), 1)

	test.IsNil(t, user.AddUserToGroup(group, member1.Name))

	members, err = user.GetMembers(group)
	test.IsNil(t, err)
	test.IsEqual(t, len(members), 2)

	members, err = user.GetMembers(group + "a")
	test.IsNotNil(t, err, "group should be invalid")
}
