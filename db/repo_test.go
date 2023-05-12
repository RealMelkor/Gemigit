package db

import (
	"testing"
	"gemigit/test"
)

const invalidRepoName = "$repo"
const validRepoName = "repo"

func TestCreateRepo(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNotNil(t, user.CreateRepo(invalidRepoName, signature),
			"name should be invalid")
	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	test.IsNotNil(t, user.CreateRepo(validRepoName, signature),
			"repository name should be already taken")
	test.IsNotNil(t, user.CreateRepo(validRepoName + "a", "invalid"),
			"signature should be invalid")
}

func TestGetRepoID(t *testing.T) {
	
	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	_, err := GetRepoID(validRepoName, user.ID)
	test.IsNil(t, err)
	_, err = GetRepoID(validRepoName + "a", user.ID)
	test.IsNotNil(t, err, "should return repository not found")
	_, err = GetRepoID(validRepoName, user.ID + 1)
	test.IsNotNil(t, err, "should return user not found")

}

func TestChangeRepoName(t *testing.T) {
	
	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))

	err := user.ChangeRepoName(validRepoName, invalidRepoName,
					signature)
	test.IsNotNil(t, err, "repo name should be invalid")

	err = user.ChangeRepoName(validRepoName, validRepoName + "a",
					signature + "a")
	test.IsNotNil(t, err, "signature should be invalid")

	err = user.ChangeRepoName(validRepoName + "a", validRepoName,
					signature)
	test.IsNotNil(t, err, "repository should be invalid")

	id, err := GetRepoID(validRepoName, user.ID)
	test.IsNil(t, err)

	test.IsNil(t, user.ChangeRepoName(validRepoName, validRepoName + "a",
					signature))

	_, err = GetRepoID(validRepoName, user.ID);
	test.IsNotNil(t, err, "should return repository not found")

	id_alt, err := GetRepoID(validRepoName + "a", user.ID)
	test.IsNil(t, err)

	test.IsEqual(t, id_alt, id)
}

func TestChangeRepoDesc(t *testing.T) {
	
	initDB(t)

	user, signature := createUserAndSession(t)
	description := test.FuncName(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))

	test.IsNotNil(t, user.ChangeRepoDesc(validRepoName + "a", description),
			"repo name should be invalid")

	test.IsNil(t, user.ChangeRepoDesc(validRepoName, description))

	repo, err := user.GetRepo(validRepoName)
	test.IsNil(t, err)
	test.IsEqual(t, repo.Description, description)
	
}

func TestGetRepoDesc(t *testing.T) {
	
	initDB(t)

	user, signature := createUserAndSession(t)
	description := test.FuncName(t)
	
	_, err := GetRepoDesc(validRepoName, user.Name)
	test.IsNotNil(t, err, "repository should be invalid")
	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	test.IsNil(t, user.ChangeRepoDesc(validRepoName, description))

	desc, err := GetRepoDesc(validRepoName, user.Name)
	test.IsNil(t, err)
	test.IsEqual(t, desc, description)
}

func TestDeleteRepo(t *testing.T) {
	
	initDB(t)

	user, signature := createUserAndSession(t)
	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	test.IsNotNil(t, user.DeleteRepo(validRepoName, signature + "a"),
			"signature should be invalid")
	test.IsNotNil(t, user.DeleteRepo(validRepoName + "a", signature),
			"repository should be invalid")
	test.IsNil(t, user.DeleteRepo(validRepoName, signature))
	_, err := user.GetRepo(validRepoName)
	test.IsNotNil(t, err, "repository should be invalid")
}

func TestIsRepoPublic(t *testing.T) {
	
	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))

	_, err := IsRepoPublic(validRepoName + "a", user.Name)
	test.IsNotNil(t, err, "repository should be invalid")

	b, err := IsRepoPublic(validRepoName, user.Name)
	test.IsNil(t, err)
	test.IsEqual(t, b, false)

	test.IsNotNil(t, user.TogglePublic(validRepoName, signature + "a"),
			"signature should be invalid")

	test.IsNotNil(t, user.TogglePublic(validRepoName + "a", signature),
			"repository should be invalid")

	test.IsNil(t, user.TogglePublic(validRepoName, signature))

	b, err = IsRepoPublic(validRepoName, user.Name)
	test.IsNil(t, err)
	test.IsEqual(t, b, true)

	test.IsNil(t, user.TogglePublic(validRepoName, signature))

	b, err = IsRepoPublic(validRepoName, user.Name)
	test.IsNil(t, err)
	test.IsEqual(t, b, false)
}

func TestGetPublicRepo(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	test.IsNil(t, user.CreateRepo(validRepoName + "a", signature))

	repos, err := GetPublicRepo()
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 0)

	test.IsNil(t, user.TogglePublic(validRepoName, signature))

	repos, err = GetPublicRepo()
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 1)

	test.IsNil(t, user.TogglePublic(validRepoName + "a", signature))

	repos, err = GetPublicRepo()
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 2)

}

func TestGetRepos(t *testing.T) {

	initDB(t)

	user, signature := createUserAndSession(t)

	test.IsNil(t, user.CreateRepo(validRepoName, signature))
	test.IsNil(t, user.CreateRepo(validRepoName + "a", signature))
	test.IsNil(t, user.TogglePublic(validRepoName, signature))

	repos, err := user.GetRepos(false)
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 2)

	repos, err = user.GetRepos(true)
	test.IsNil(t, err)
	test.IsEqual(t, len(repos), 1)
}
