package gmi

import (
	"bytes"
	"errors"
	"fmt"
	"gemigit/access"
	"gemigit/config"
	"gemigit/db"
	"gemigit/repo"
	io "gemigit/util"
	"log"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pitr/gig"
)

func execT(c gig.Context, template string, data interface{}) error {
	t := templates.Lookup(template)
	var b bytes.Buffer
	err := t.Execute(&b, data)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure, err.Error())
	}
	return c.Gemini(b.String())
}

func execTemplate(template string, data interface{}) (string, error) {
	t := templates.Lookup(template)
	var b bytes.Buffer
	err := t.Execute(&b, data)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

const (
	pageLog = iota
	pageFiles
	pageRefs
	pageLicense
	pageReadme
)

var templates *template.Template

func LoadTemplate(dir string) error {
	var err error

	dirlen := len(dir)
	if dirlen > 1 && dir[dirlen - 1] == '/' {
		dir = dir[:dirlen - 1]
	}

	templates = template.New("gmi")
	template.Must(templates.Funcs(template.FuncMap {
		"AccessFirst": accessFirstOption,
		"AccessSecond": accessSecondOption,
		"AccessPrivilege": privilegeToString,
	}).ParseFiles(
		dir + "/index.gmi",
		dir + "/account.gmi",
		dir + "/repo.gmi",
		dir + "/repo_log.gmi",
		dir + "/repo_files.gmi",
		dir + "/repo_refs.gmi",
		dir + "/repo_license.gmi",
		dir + "/repo_readme.gmi",
		dir + "/repo_access.gmi",
		dir + "/register_success.gmi",
		dir + "/public_repo.gmi",
		dir + "/group_list.gmi",
		dir + "/group.gmi",
		dir + "/public_list.gmi",
		dir + "/public_user.gmi",
	))
	if err != nil {
		return err
	}
	log.Println("Templates loaded")
	return nil
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

func ShowIndex(c gig.Context) (error) {
	_, connected := db.GetUser(c.CertHash())
	data := struct {
		Title string
		Registration bool
		Connected bool
		Public bool
	}{
		Title: config.Cfg.Title,
		Registration: config.Cfg.Users.Registration,
		Connected: connected,
		Public: connected || config.Cfg.Git.Public,
	}
	return execT(c, "index.gmi", data)
}

func ShowAccount(c gig.Context) (error) {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid username")
	}
	repoNames := []string{}
	repos, err := user.GetRepos(false)
	if err != nil {
		repoNames = []string{"Failed to load repositories"}
		log.Println(err)
	} else {
		for _, repo := range repos {
			repoNames = append(repoNames, repo.Name)
		}
	}
	accessRepos, err := db.HasReadAccessTo(user.ID)
	data := struct {
		Username string
		Description string
		Repositories []string
		RepositoriesAccess []db.Repo
	}{
		Username: user.Name,
		Description: user.Description,
		Repositories: repoNames,
		RepositoriesAccess: accessRepos,
	}
	return execT(c, "account.gmi", data)
}

func ShowGroups(c gig.Context) (error) {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid username")
	}
	groups, err := user.GetGroups()
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Failed to fetch groups")
	}
	data := struct {
		Groups []db.Group
	}{
		Groups: groups,
	}
	return execT(c, "group_list.gmi", data)
}

func ShowMembers(c gig.Context) (error) {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid username")
	}
	group := c.Param("group")
	isOwner, err := user.IsInGroup(group)
	if err != nil {
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Group not found")
	}

	members, err := user.GetMembers(group)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Failed to fetch group members")
	}
	desc, err := db.GetGroupDesc(group)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Failed to fetch group description")
	}

	owner := ""
	if isOwner {
		owner = user.Name
	} else {
		m, err := db.GetGroupOwner(group)
		if err != nil {
			log.Println(err.Error())
			return c.NoContent(gig.StatusTemporaryFailure,
					   "Failed to fetch group owner")
		}
		owner = m.Name
	}

	data := struct {
		Members []db.Member
		MembersCount int
		IsOwner bool
		Owner string
		Group string
		Description string
	}{
		Members: members,
		MembersCount: len(members),
		IsOwner: isOwner,
		Owner: owner,
		Group: group,
		Description: desc,
	}
	return execT(c, "group.gmi", data)
}

func getRepo(c gig.Context, owner bool) (string, string, error) {
        username := ""
        if owner {
                user, exist := db.GetUser(c.CertHash())
                if !exist {
                        return "", "", c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
                }
                username = user.Name
        } else {
                username = c.Param("user")
                ret, err := db.IsRepoPublic(c.Param("repo"), c.Param("user"))
		if !ret {
			user, exist := db.GetUser(c.CertHash())
			if exist {
				err := access.HasReadAccess(c.Param("repo"),
							    c.Param("user"),
							    user.Name)
				ret = err == nil
			}
		}
                if !ret || err != nil {
                        return "", "", c.NoContent(gig.StatusBadRequest,
				"No repository called " + c.Param("repo") +
                                " by user " + c.Param("user"))
                }
        }
	return username, c.Param("repo"), nil
}

func hasFile(name string, author string, file string) bool {
	ret, err := repo.GetFile(name, author, file)
	if ret != nil && err == nil {
		return true
	} 
	return false
}

type commit struct {
	Message string
	Info string
}

type file struct {
	Hash string
	Info string
}

type branch struct {
	Name string
	Info string
}

func showRepoLogs(name string, author string) (string, error) {
	ret, err := repo.GetCommits(name, author)
	if ret == nil || err == transport.ErrEmptyRemoteRepository {
		return "", nil
	}
	if err != nil {
		log.Println(err.Error())
		return "", errors.New("Corrupted repository")
	}
	commits := []commit{}
	err = ret.ForEach(func(c *object.Commit) error {
		info := c.Hash.String() + ", by " + c.Author.Name + " on " +
			c.Author.When.Format("2006-01-02 15:04:05")
		commits = append(commits, commit{Info: info,
						 Message: c.Message})
		return nil
	})
	return execTemplate("repo_log.gmi", commits)
}

func showRepoFiles(name string, author string) (string, error) {
	ret, err := repo.GetFiles(name, author)
	if ret == nil || err == transport.ErrEmptyRemoteRepository {
		return "", nil
	}
	if err != nil {
		log.Println(err.Error())
		return "", errors.New("Corrupted repository")
	}
	files := []file{}
	err = ret.ForEach(func(f *object.File) error {
		info := f.Mode.String() + " " + f.Name +
			" " + strconv.Itoa(int(f.Size))
		files = append(files, file{Info: info,
					   Hash: f.Blob.Hash.String()})
		return nil
	})
	return execTemplate("repo_files.gmi", files)
}

func showRepoRefs(name string, author string) (string, error) {
	refs, err := repo.GetRefs(name, author)
	if refs == nil || err == transport.ErrEmptyRemoteRepository {
		return "", nil
	}
	if err != nil {
		log.Println(err)
		return "", errors.New("Corrupted repository")
	}
	branches := []branch{}
	tags := []branch{}
	err = refs.ForEach(func(c *plumbing.Reference) error {
		if c.Type().String() != "hash-reference" ||
		   c.Name().IsRemote() {
			return nil
		}
		var b branch
		b.Name = c.Name().String()
		b.Name = b.Name[strings.LastIndex(b.Name, "/") + 1:]
		b.Info = "last commit on "

		commit, err := repo.GetCommit(name, author, c.Hash())
		if err != nil {
			b.Info = "failed to fetch commit"
		} else {
			when := commit.Author.When
			str := fmt.Sprintf(
				"%d-%02d-%02d %02d:%02d:%02d",
				when.Year(), int(when.Month()),
				when.Day(), when.Hour(),
				when.Minute(), when.Second())
			b.Info += str + " by " + commit.Author.Name
		}
		if c.Name().IsBranch() {
			branches = append(branches, b)
		} else {
			tags = append(tags, b)
		}
		return nil
	})
	refs.Close()
	data := struct {
		Branches []branch
		Tags []branch
	}{
		branches,
		tags,
	}
	return execTemplate("repo_refs.gmi", data)
}

func showRepoLicense(name string, author string) (string, error) {
	content, err := showRepoFile(author, name, "LICENSE")
	if err != nil {
		return "", errors.New("No license found")
	}
	return execTemplate("repo_license.gmi", content)
}

func showRepoReadme(name string, author string) (string, error) {
	content, err := showRepoFile(author, name, "README.gmi")
	if err != nil {
		content, err = showRepoFile(author, name, "README")
	}
	if err != nil {
		return "", errors.New("No readme found")
	}
	return execTemplate("repo_readme.gmi", content)
}

func showRepo(c gig.Context, page int, owner bool) (error) {
	author, name, err := getRepo(c, owner)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure, err.Error())
	}
	desc, err := db.GetRepoDesc(name, author)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Repository not found")
	}
	protocol := "http"
	if config.Cfg.Git.Https {
		protocol = "https"
	}
	public, err := db.IsRepoPublic(name, author)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Repository not found")
	}

	content := ""
	switch page {
	case pageLog:
		content, err = showRepoLogs(name, author)
	case pageFiles:
		content, err = showRepoFiles(name, author)
	case pageRefs:
		content, err = showRepoRefs(name, author)
	case pageLicense:
		content, err = showRepoLicense(name, author)
	case pageReadme:
		content, err = showRepoReadme(name, author)
	}
	if err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid repository")
	}
	
	data := struct {
		Protocol string
		Domain string
		User string
		Description string
		Repo string
		Public bool
		HasReadme bool
		HasLicense bool
		Content string
	}{
		Protocol: protocol,
		Domain: config.Cfg.Git.Domain,
		User: author,
		Description: desc,
		Repo: name,
		Public: public,
		HasReadme: hasFile(name, author, "README.gmi") ||
			   hasFile(name, author, "README"),
		HasLicense: hasFile(name, author, "LICENSE"),
		Content: content,
	}
	if owner {
		return execT(c, "repo.gmi", data)
	}
	return execT(c, "public_repo.gmi", data)
}

func PublicList(c gig.Context) (error) {
	repos, err := db.GetPublicRepo()
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Internal error, "+err.Error())
	}
	return execT(c, "public_list.gmi", repos)
}

func PublicAccount(c gig.Context) error {
	user, err := db.GetPublicUser(c.Param("user"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	repos, err := user.GetRepos(true)
	if err != nil {
		return c.NoContent(gig.StatusTemporaryFailure,
				   "Invalid account, " + err.Error())
	}
	data := struct {
		Name string
		Description string
		Repositories []db.Repo
	}{
		user.Name,
		user.Description,
		repos,
	}
	return execT(c, "public_user.gmi", data)
}

func ShowAccess(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid username")
	}
	repo, err := user.GetRepo(c.Param("repo"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	access, err := db.GetRepoAccess(repo.RepoID)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	groups, err := db.GetRepoGroupAccess(repo.RepoID)
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	data := struct {
		Repo string
		Collaborators []db.Access
		Groups []db.Access
		Owner bool
	}{
		Repo: repo.Name,
		Collaborators: access,
		Groups: groups,
		Owner: true,
	}
	return execT(c, "repo_access.gmi", data)
}
