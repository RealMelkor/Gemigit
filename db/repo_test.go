package db

import (
	"testing"
)

const invalidRepoName = "$repo"
const validRepoName = "repo"

func TestCreateRepo(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)

	isNotNil(t, user.CreateRepo(invalidRepoName, signature),
			"name should be invalid")
	isNil(t, user.CreateRepo(validRepoName, signature))
	isNotNil(t, user.CreateRepo(validRepoName, signature),
			"repository name should be already taken")
	isNotNil(t, user.CreateRepo(validRepoName + "a", "invalid"),
			"signature should be invalid")
}

func TestGetRepoID(t *testing.T) {
	
	initDB(t)

	user, _, signature := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	_, err := GetRepoID(validRepoName, user.ID)
	isNil(t, err)
	_, err = GetRepoID(validRepoName + "a", user.ID)
	isNotNil(t, err, "should return repository not found")
	_, err = GetRepoID(validRepoName, user.ID + 1)
	isNotNil(t, err, "should return user not found")

}

func TestChangeRepoName(t *testing.T) {
	
	initDB(t)

	user, _, signature := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))

	err := user.ChangeRepoName(validRepoName, invalidRepoName,
					signature)
	isNotNil(t, err, "repo name should be invalid")

	err = user.ChangeRepoName(validRepoName, validRepoName + "a",
					signature + "a")
	isNotNil(t, err, "signature should be invalid")

	err = user.ChangeRepoName(validRepoName + "a", validRepoName,
					signature)
	isNotNil(t, err, "repository should be invalid")

	id, err := GetRepoID(validRepoName, user.ID)
	isNil(t, err)

	isNil(t, user.ChangeRepoName(validRepoName, validRepoName + "a",
					signature))

	_, err = GetRepoID(validRepoName, user.ID);
	isNotNil(t, err, "should return repository not found")

	id_alt, err := GetRepoID(validRepoName + "a", user.ID)
	isNil(t, err)

	isEqual(t, id_alt, id)
}

func TestChangeRepoDesc(t *testing.T) {
	
	initDB(t)

	user, _, signature := createUserAndSession(t)
	description := funcName(t)

	isNil(t, user.CreateRepo(validRepoName, signature))

	isNotNil(t, user.ChangeRepoDesc(validRepoName + "a", description),
			"repo name should be invalid")

	isNil(t, user.ChangeRepoDesc(validRepoName, description))

	repo, err := user.GetRepo(validRepoName)
	isNil(t, err)
	isEqual(t, repo.Description, description)
	
}

func TestGetRepoDesc(t *testing.T) {
	
	initDB(t)

	user, _, signature := createUserAndSession(t)
	description := funcName(t)
	
	_, err := GetRepoDesc(validRepoName, user.Name)
	isNotNil(t, err, "repository should be invalid")
	isNil(t, user.CreateRepo(validRepoName, signature))
	isNil(t, user.ChangeRepoDesc(validRepoName, description))

	desc, err := GetRepoDesc(validRepoName, user.Name)
	isNil(t, err)
	isEqual(t, desc, description)
}

func TestDeleteRepo(t *testing.T) {
	
	initDB(t)

	user, _, signature := createUserAndSession(t)
	isNil(t, user.CreateRepo(validRepoName, signature))
	isNotNil(t, user.DeleteRepo(validRepoName, signature + "a"),
			"signature should be invalid")
	isNotNil(t, user.DeleteRepo(validRepoName + "a", signature),
			"repository should be invalid")
	isNil(t, user.DeleteRepo(validRepoName, signature))
	_, err := user.GetRepo(validRepoName)
	isNotNil(t, err, "repository should be invalid")
}

func TestIsRepoPublic(t *testing.T) {
	
	initDB(t)

	user, _, signature := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))

	_, err := IsRepoPublic(validRepoName + "a", user.Name)
	isNotNil(t, err, "repository should be invalid")

	b, err := IsRepoPublic(validRepoName, user.Name)
	isNil(t, err)
	isEqual(t, b, false)

	isNotNil(t, user.TogglePublic(validRepoName, signature + "a"),
			"signature should be invalid")

	isNotNil(t, user.TogglePublic(validRepoName + "a", signature),
			"repository should be invalid")

	isNil(t, user.TogglePublic(validRepoName, signature))

	b, err = IsRepoPublic(validRepoName, user.Name)
	isNil(t, err)
	isEqual(t, b, true)

	isNil(t, user.TogglePublic(validRepoName, signature))

	b, err = IsRepoPublic(validRepoName, user.Name)
	isNil(t, err)
	isEqual(t, b, false)
}

func TestGetPublicRepo(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	isNil(t, user.CreateRepo(validRepoName + "a", signature))

	repos, err := GetPublicRepo()
	isNil(t, err)
	isEqual(t, len(repos), 0)

	isNil(t, user.TogglePublic(validRepoName, signature))

	repos, err = GetPublicRepo()
	isNil(t, err)
	isEqual(t, len(repos), 1)

	isNil(t, user.TogglePublic(validRepoName + "a", signature))

	repos, err = GetPublicRepo()
	isNil(t, err)
	isEqual(t, len(repos), 2)

}

func TestGetRepos(t *testing.T) {

	initDB(t)

	user, _, signature := createUserAndSession(t)

	isNil(t, user.CreateRepo(validRepoName, signature))
	isNil(t, user.CreateRepo(validRepoName + "a", signature))
	isNil(t, user.TogglePublic(validRepoName, signature))

	repos, err := user.GetRepos(false)
	isNil(t, err)
	isEqual(t, len(repos), 2)

	repos, err = user.GetRepos(true)
	isNil(t, err)
	isEqual(t, len(repos), 1)
}
