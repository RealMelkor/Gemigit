package main

import (
	"gemigit/db"
	"gemigit/repo"
	"io"
	"log"
	"strconv"

	"gemigit/config"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pitr/gig"
)

func getHttpAddress(user string, repo string) string {
	ret := "git clone "
	if config.Cfg.Gemigit.Https {
		ret += "https://"
	} else {
		ret += "http://"
	}
	ret += config.Cfg.Gemigit.Domain + "/" + user + "/" + repo + "\n"
	return ret
}

func showRepoHeader(user string, reponame string, owner bool) (string, error) {
	ret := ""
	if owner {
		ret += "=>/account Go back\n\n"
	} else {
		ret += "=>/repo Go back\n\n"
	}
	ret += "# " + reponame
	if !owner {
		ret += " by " + user + "\n=>/repo/" + user + " View account\n"
	} else {
		ret += "\n"
	}
	desc, err := db.GetRepoDesc(reponame, user)
	if err != nil {
		return "", err
	}
	if desc != "" {
		ret += "> " + desc + "\n"
	}
	ret += "> " + getHttpAddress(user, reponame)
	if owner {
		ret += "\n" +
		"=>/account/repo/" + reponame + 
		"/chname Change repository name\n" +
		"=>/account/repo/" + reponame + 
		"/chdesc Change repository description\n" +
		"=>/account/repo/" + reponame + 
		"/delrepo Delete repository\n" +
		"=>/account/repo/" + reponame + 
		"/togglepublic Make the repository "
		b, err := db.IsRepoPublic(reponame, user)
		if err != nil {
			return "", err
		}
		if b {
			ret += "private\n\n"
		} else {
			ret += "public\n\n"
		}
		ret += "=>/account/repo/" + reponame + " Log\n"
		ret += "=>/account/repo/" + reponame + "/files Files\n"
		file, err := repo.GetFile(reponame, user, "LICENSE")
		if file != nil && err == nil {
			ret += "=>/account/repo/" + reponame + "/license License\n"
		}
		file, err = repo.GetFile(reponame, user, "README")
		if file != nil && err == nil {
			ret += "=>/account/repo/" + reponame + "/readme Readme\n"
		}
	} else {
		ret += "\n=>/repo/" + user + "/" + reponame + " Log\n"
		ret += "=>/repo/" + user + "/" + reponame + "/files Files\n"
		file, err := repo.GetFile(reponame, user, "LICENSE")
		if file != nil && err == nil {
			ret += "=>/repo/" + user + "/" + 
				reponame + "/license License\n"
		}
		file, err = repo.GetFile(reponame, user, "README")
		if file != nil && err == nil {
			ret += "=>/repo/" + user + "/" + 
				reponame + "/readme Readme\n"
		}
	}

	return ret, nil
}

func showRepoFiles(user string, reponame string, owner bool) (string, error) {
	files, err := repo.GetFiles(reponame, user)
	if err != nil {
		log.Println(err.Error())
	}
	ret := "\n## Files\n\n"
	if files != nil {
		err = files.ForEach(func(f *object.File) error {
			if owner {
				ret += "=>/account/repo/"
			} else {
				ret += "=>/repo/" + user + "/"
			}
			ret += reponame + "/files/" + 
			       f.Blob.Hash.String() + " " + 
			       f.Mode.String() + " " + 
			       f.Name + " " + 
			       strconv.Itoa(int(f.Size)) + "\n"
			return nil
		})
		if err != nil {
			return "", err
		}
	} else {
		ret += "Empty repository\n"
	}
	return ret, nil
}

func showRepoCommits(user string, reponame string) (string, error) {
	commits, err := repo.GetCommits(reponame, user)
	if err != nil {
		return "", err
	}
	ret := "\n## Commits\n\n"
	if commits != nil {
		err = commits.ForEach(func(c *object.Commit) error {
			ret += "* " + c.Hash.String() + 
				", by " + c.Author.Name +
				" on " + c.Author.When.Format("2006-01-02 15:04:05") + 
				"\n"
			ret += "> " + c.Message + "\n"
			return nil
		})
		if err != nil {
			return "", err
		}
	} else {
		ret += "Empty repository\n"
	}
	return ret, nil
}

func showRepoFile(user string, reponame string, file string) (string, error) {
	out, err := repo.GetFile(reponame, user, file)
	if err != nil {
		return "", err
	}
	reader, err := out.Reader()
	if err != nil {
		return "", err
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func showLicense(user string, reponame string) (string, error) {
	ret := "\n## License\n\n"
	file, err := showRepoFile(user, reponame, "LICENSE")
	if err != nil {
		return "", err
	}
	ret += file
	return ret, nil
}

func showReadme(user string, reponame string) (string, error) {
	ret := "\n## Readme\n\n"
	file, err := showRepoFile(user, reponame, "README")
	if err != nil {
		return "", err
	}
	ret += file
	return ret, nil
}

func repoRequest(c gig.Context, param string, owner bool) error {
	username := ""
	if owner {
		user, exist := db.GetUser(c.CertHash())
		if !exist {
			return c.NoContent(gig.StatusBadRequest, "Invalid username")
		}
		username = user.Name
	} else {
		username = c.Param("user")
	}
	ret, err := showRepoHeader(username, c.Param("repo"), owner)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}

	var buf string
	switch param {
	case "":
		buf, err = showRepoCommits(username, c.Param("repo"))
	case "files":
		buf, err = showRepoFiles(username, c.Param("repo"), owner)
	case "license":
		buf, err = showLicense(username, c.Param("repo"))
	case "readme":
		buf, err = showReadme(username, c.Param("repo"))
	}
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	ret += buf
	return c.Gemini(ret)
}
