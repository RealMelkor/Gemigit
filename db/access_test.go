package db

import "testing"

func TestAddUserAccess(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)
	user2, _, signature2 := createUserAndSession(t)

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

	user, _, signature := createUserAndSession(t)
	user2, _, signature2 := createUserAndSession(t)

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

	user, _, signature := createUserAndSession(t)
	user2, _, _ := createUserAndSession(t)

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

	user, _, signature := createUserAndSession(t)
	user2, _, _ := createUserAndSession(t)

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
