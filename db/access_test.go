package db

import "testing"

func TestAddUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, signature2 := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	isNil(t, user2.CreateRepo(validRepoName, signature2))
	repo2, err := user2.GetRepo(validRepoName)
	isNil(t, err)

	isNil(t, user.AddUserAccess(repo, user2.Name))
	isNotNil(t, user.AddUserAccess(repo, user2.Name + "a"), "invalid user")
	isNotNil(t, user.AddUserAccess(repo2, user2.Name),
			"not the owner of the repository")
	isNotNil(t, user.AddUserAccess(repo, user.Name),
			"the owner already has access")
	isNotNil(t, user.AddUserAccess(repo, user2.Name),
			"user already has access")

}

func TestAddGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, signature2 := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	isNil(t, user2.CreateRepo(validRepoName, signature2))
	repo2, err := user2.GetRepo(validRepoName)
	isNil(t, err)

	group := funcName(t)
	group2 := group + "2"
	isNil(t, user.CreateGroup(group, signature))
	isNil(t, user2.CreateGroup(group2, signature2))

	isNil(t, user.AddGroupAccess(repo, group))
	isNotNil(t, user.AddGroupAccess(repo, group + "a"), "invalid group")
	isNotNil(t, user.AddGroupAccess(repo2, group2),
			"not the owner of the repository")
	isNotNil(t, user.AddGroupAccess(repo, group),
			"group already has access")

}

func TestRemoveUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	isNil(t, user.AddUserAccess(repo, user2.Name))
	isNotNil(t, user2.RemoveUserAccess(repo, user2.ID),
			"not the repository owner")
	isNil(t, user.RemoveUserAccess(repo, user2.ID))
	isNotNil(t, user.RemoveUserAccess(repo, user2.ID),
			"access already revoked")

}

func TestRemoveGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	group := funcName(t)
	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, user.AddGroupAccess(repo, group))
	isNotNil(t, user2.RemoveGroupAccess(repo, id),
			"not the repository owner")
	isNil(t, user.RemoveGroupAccess(repo, id))
	isNotNil(t, user.RemoveGroupAccess(repo, id),
			"access already revoked")

}

func TestSetUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	isNil(t, user.AddUserAccess(repo, user2.Name))

	isNil(t, user.SetUserAccess(repo, user2.ID, 2))
	isNotNil(t, user.SetUserAccess(repo, user.ID, 2),
			"cannot change owner access")
	isNotNil(t, user2.SetUserAccess(repo, user2.ID, 2),
			"only the owner can manage access")

}

func TestSetGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	group := funcName(t)
	isNil(t, user.CreateGroup(group, signature))

	isNil(t, user.AddGroupAccess(repo, group))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, user.SetGroupAccess(repo, id, 2))
	isNotNil(t, user2.SetGroupAccess(repo, id, 2),
			"only the owner can manage access")

}

func TestGetUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	access, err := GetUserAccess(repo, user2)
	isNil(t, err)
	isEqual(t, access, accessNone)

	access, err = GetUserAccess(repo, user)
	isNil(t, err)
	isEqual(t, access, accessMax)

	isNil(t, user.AddUserAccess(repo, user2.Name))

	isNil(t, user.SetUserAccess(repo, user2.ID, accessRead))
	access, err = GetUserAccess(repo, user2)
	isNil(t, err)
	isEqual(t, access, accessRead)

	isNil(t, user.SetUserAccess(repo, user2.ID, accessReadWrite))
	access, err = GetUserAccess(repo, user2)
	isNil(t, err)
	isEqual(t, access, accessReadWrite)

}

func TestGetGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	group := funcName(t)
	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	_, err = GetGroupAccess(repo, id)
	isNotNil(t, err, "the group does not have any access")

	isNil(t, user.AddGroupAccess(repo, group))

	isNil(t, user.SetGroupAccess(repo, id, 1))
	access, err := GetGroupAccess(repo, id)
	isNil(t, err)
	isEqual(t, access, 1)

	isNil(t, user.SetGroupAccess(repo, id, 0))
	access, err = GetGroupAccess(repo, id)
	isNil(t, err)
	isEqual(t, access, 0)

}

func TestGetUserGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	group := funcName(t)
	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, user.AddGroupAccess(repo, group))

	access, err := GetUserGroupAccess(repo, user2)
	isNil(t, err)
	isEqual(t, access, -1)

	isNil(t, user.AddUserToGroup(group, user2.Name))

	access, err = GetUserGroupAccess(repo, user2)
	isNil(t, err)
	isEqual(t, access, accessRead)

	isNil(t, user.SetGroupAccess(repo, id, accessReadWrite))

	access, err = GetUserGroupAccess(repo, user2)
	isNil(t, err)
	isEqual(t, access, accessReadWrite)

}

func TestGetAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	group := funcName(t)
	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, user.AddGroupAccess(repo, group))

	access, err := GetAccess(user, repo)
	isNil(t, err)
	isEqual(t, access, accessMax)

	access, err = GetAccess(user2, repo)
	isNil(t, err)
	isEqual(t, access, accessNone)

	isNil(t, user.AddUserToGroup(group, user2.Name))

	access, err = GetAccess(user2, repo)
	isNil(t, err)
	isEqual(t, access, accessRead)

	isNil(t, user.SetGroupAccess(repo, id, 2))

	access, err = GetAccess(user2, repo)
	isNil(t, err)
	isEqual(t, access, 2)

}

func TestGetRepoUserAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, _ := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	access, err := GetRepoUserAccess(repo.ID)
	isNil(t, err)
	isEqual(t, len(access), 0)

	isNil(t, user.AddUserAccess(repo, user2.Name))

	access, err = GetRepoUserAccess(repo.ID)
	isNil(t, err)
	isEqual(t, len(access), 1)
	isEqual(t, access[0].UserID, user2.ID)
	isEqual(t, access[0].Privilege, accessDefault)

	isNil(t, user.SetUserAccess(repo, user2.ID, accessReadWrite))

	access, err = GetRepoUserAccess(repo.ID)
	isNil(t, err)
	isEqual(t, len(access), 1)
	isEqual(t, access[0].UserID, user2.ID)
	isEqual(t, access[0].Privilege, accessReadWrite)

}

func TestGetRepoGroupAccess(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	group := funcName(t)
	isNil(t, user.CreateGroup(group, signature))
	id, err := GetGroupID(group)
	isNil(t, err)

	isNil(t, user.CreateRepo(validRepoName, signature))
	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)

	access, err := GetRepoGroupAccess(repo.ID)
	isNil(t, err)
	isEqual(t, len(access), 0)

	isNil(t, user.AddGroupAccess(repo, group))

	access, err = GetRepoGroupAccess(repo.ID)
	isNil(t, err)
	isEqual(t, len(access), 1)
	isEqual(t, access[0].GroupID, id)
	isEqual(t, access[0].Privilege, accessDefault)

	isNil(t, user.SetGroupAccess(repo, id, accessReadWrite))

	access, err = GetRepoGroupAccess(repo.ID)
	isNil(t, err)
	isEqual(t, len(access), 1)
	isEqual(t, access[0].GroupID, id)
	isEqual(t, access[0].Privilege, accessReadWrite)

}

func TestHasReadAccessTo(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)
	user2, signature2 := createUserAndSession(t)

	repos, err := user.HasReadAccessTo()
	isNil(t, err)
	isEqual(t, len(repos), 0)

	isNil(t, user2.CreateRepo(validRepoName, signature2))
	repo, err := user2.GetRepo(validRepoName)
	isNil(t, err)

	isNil(t, user2.CreateRepo(validRepoName + "2", signature2))
	repo2, err := user2.GetRepo(validRepoName + "2")
	isNil(t, err)

	group := funcName(t)
	isNil(t, user.CreateGroup(group, signature))

	user2.AddUserAccess(repo, user.Name)

	repos, err = user.HasReadAccessTo()
	isNil(t, err)
	isEqual(t, len(repos), 1)

	user2.AddGroupAccess(repo2, group)

	repos, err = user.HasReadAccessTo()
	isNil(t, err)
	isEqual(t, len(repos), 2)



}
