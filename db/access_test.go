package db

import (
	"testing"
	"gemigit/test"
)

func TestAddUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, signature2 := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	test.IsNil(t, user2.CreateRepo(validRepoName, signature2))
	repo2, err := user2.GetRepo(validRepoName)
	test.IsNil(t, err)

	test.IsNil(t, user.AddUserAccess(repo, user2.Name))
	test.IsNotNil(t, user.AddUserAccess(repo, user2.Name + "a"),
			"invalid user")
	test.IsNotNil(t, user.AddUserAccess(repo2, user2.Name),
			"not the owner of the repository")
	test.IsNotNil(t, user.AddUserAccess(repo, user.Name),
			"the owner already has access")
	test.IsNotNil(t, user.AddUserAccess(repo, user2.Name),
			"user already has access")

}

func TestAddGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, signature2 := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	test.IsNil(t, user2.CreateRepo(validRepoName, signature2))
	repo2, err := user2.GetRepo(validRepoName)
	test.IsNil(t, err)

	group := test.FuncName(t)
	group2 := group + "2"
	test.IsNil(t, user.CreateGroup(group, signature))
	test.IsNil(t, user2.CreateGroup(group2, signature2))

	test.IsNil(t, user.AddGroupAccess(repo, group))
	test.IsNotNil(t, user.AddGroupAccess(repo, group + "a"),
			"invalid group")
	test.IsNotNil(t, user.AddGroupAccess(repo2, group2),
			"not the owner of the repository")
	test.IsNotNil(t, user.AddGroupAccess(repo, group),
			"group already has access")

}

func TestRemoveUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	test.IsNil(t, user.AddUserAccess(repo, user2.Name))
	test.IsNotNil(t, user2.RemoveUserAccess(repo, user2.ID),
			"not the repository owner")
	test.IsNil(t, user.RemoveUserAccess(repo, user2.ID))
	test.IsNotNil(t, user.RemoveUserAccess(repo, user2.ID),
			"access already revoked")

}

func TestRemoveGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	group := test.FuncName(t)
	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.AddGroupAccess(repo, group))
	test.IsNotNil(t, user2.RemoveGroupAccess(repo, id),
			"not the repository owner")
	test.IsNil(t, user.RemoveGroupAccess(repo, id))
	test.IsNotNil(t, user.RemoveGroupAccess(repo, id),
			"access already revoked")

}

func TestSetUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	test.IsNil(t, user.AddUserAccess(repo, user2.Name))

	test.IsNil(t, user.SetUserAccess(repo, user2.ID, 2))
	test.IsNotNil(t, user.SetUserAccess(repo, user.ID, 2),
			"cannot change owner access")
	test.IsNotNil(t, user2.SetUserAccess(repo, user2.ID, 2),
			"only the owner can manage access")

}

func TestSetGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	group := test.FuncName(t)
	test.IsNil(t, user.CreateGroup(group, signature))

	test.IsNil(t, user.AddGroupAccess(repo, group))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.SetGroupAccess(repo, id, 2))
	test.IsNotNil(t, user2.SetGroupAccess(repo, id, 2),
			"only the owner can manage access")

}

func TestGetUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	access, err := GetUserAccess(repo, user2)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessNone)

	access, err = GetUserAccess(repo, user)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessMax)

	test.IsNil(t, user.AddUserAccess(repo, user2.Name))

	test.IsNil(t, user.SetUserAccess(repo, user2.ID, accessRead))
	access, err = GetUserAccess(repo, user2)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessRead)

	test.IsNil(t, user.SetUserAccess(repo, user2.ID, accessReadWrite))
	access, err = GetUserAccess(repo, user2)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessReadWrite)

}

func TestGetGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	group := test.FuncName(t)
	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	_, err = GetGroupAccess(repo, id)
	test.IsNotNil(t, err, "the group does not have any access")

	test.IsNil(t, user.AddGroupAccess(repo, group))

	test.IsNil(t, user.SetGroupAccess(repo, id, 1))
	access, err := GetGroupAccess(repo, id)
	test.IsNil(t, err)
	test.IsEqual(t, access, 1)

	test.IsNil(t, user.SetGroupAccess(repo, id, 0))
	access, err = GetGroupAccess(repo, id)
	test.IsNil(t, err)
	test.IsEqual(t, access, 0)

}

func TestGetUserGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	group := test.FuncName(t)
	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.AddGroupAccess(repo, group))

	access, err := GetUserGroupAccess(repo, user2)
	test.IsNil(t, err)
	test.IsEqual(t, access, -1)

	test.IsNil(t, user.AddUserToGroup(group, user2.Name))

	access, err = GetUserGroupAccess(repo, user2)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessRead)

	test.IsNil(t, user.SetGroupAccess(repo, id, accessReadWrite))

	access, err = GetUserGroupAccess(repo, user2)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessReadWrite)

}

func TestGetAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	group := test.FuncName(t)
	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.AddGroupAccess(repo, group))

	access, err := GetAccess(user, repo)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessMax)

	access, err = GetAccess(user2, repo)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessNone)

	test.IsNil(t, user.AddUserToGroup(group, user2.Name))

	access, err = GetAccess(user2, repo)
	test.IsNil(t, err)
	test.IsEqual(t, access, accessRead)

	test.IsNil(t, user.SetGroupAccess(repo, id, 2))

	access, err = GetAccess(user2, repo)
	test.IsNil(t, err)
	test.IsEqual(t, access, 2)

}

func TestGetRepoUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	access, err := GetRepoUserAccess(repo.ID)
	test.IsNil(t, err)
	test.IsEqual(t, len(access), 0)

	test.IsNil(t, user.AddUserAccess(repo, user2.Name))

	access, err = GetRepoUserAccess(repo.ID)
	test.IsNil(t, err)
	test.IsEqual(t, len(access), 1)
	test.IsEqual(t, access[0].UserID, user2.ID)
	test.IsEqual(t, access[0].Privilege, accessDefault)

	test.IsNil(t, user.SetUserAccess(repo, user2.ID, accessReadWrite))

	access, err = GetRepoUserAccess(repo.ID)
	test.IsNil(t, err)
	test.IsEqual(t, len(access), 1)
	test.IsEqual(t, access[0].UserID, user2.ID)
	test.IsEqual(t, access[0].Privilege, accessReadWrite)

}

func TestGetRepoGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	group := test.FuncName(t)
	test.IsNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	test.IsNil(t, err)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)

	access, err := GetRepoGroupAccess(repo.ID)
	test.IsNil(t, err)
	test.IsEqual(t, len(access), 0)

	test.IsNil(t, user.AddGroupAccess(repo, group))

	access, err = GetRepoGroupAccess(repo.ID)
	test.IsNil(t, err)
	test.IsEqual(t, len(access), 1)
	test.IsEqual(t, access[0].GroupID, id)
	test.IsEqual(t, access[0].Privilege, accessDefault)

	test.IsNil(t, user.SetGroupAccess(repo, id, accessReadWrite))

	access, err = GetRepoGroupAccess(repo.ID)
	test.IsNil(t, err)
	test.IsEqual(t, len(access), 1)
	test.IsEqual(t, access[0].GroupID, id)
	test.IsEqual(t, access[0].Privilege, accessReadWrite)

}

func TestHasReadAccessTo(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, signature2 := createUserAndSession(t)

	repos, err := user.HasReadAccessTo()
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 0)

	test.IsNil(t, user2.CreateRepo(validRepoName, signature2))
	repo, err := user2.GetRepo(validRepoName)
	test.IsNil(t, err)

	test.IsNil(t, user2.CreateRepo(validRepoName + "2", signature2))
	repo2, err := user2.GetRepo(validRepoName + "2")
	test.IsNil(t, err)

	group := test.FuncName(t)
	test.IsNil(t, user.CreateGroup(group, signature))

	user2.AddUserAccess(repo, user.Name)

	repos, err = user.HasReadAccessTo()
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 1)

	user2.AddGroupAccess(repo2, group)

	repos, err = user.HasReadAccessTo()
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 2)
}
