package main

import (
	"bytes"
	"io"
	"fmt"
	"strconv"
	"strings"
	"gemigit/config"
	"gemigit/db"
	"gemigit/repo"
	"log"
	"text/template"

	"github.com/pitr/gig"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	pageLog = iota
	pageFiles
	pageRefs
	pageLicense
	pageReadme
)

var mainPage *template.Template
var accountPage *template.Template
var repoPage *template.Template
var repoPublicPage *template.Template
var groupListPage *template.Template
var groupMembersPage *template.Template

func loadTemplate() error {
	var err error
	mainPage, err = template.ParseFiles("templates/index.gmi")
	if err != nil {
		return err
	}
	accountPage, err = template.ParseFiles("templates/account.gmi")
	if err != nil {
		return err
	}
	repoPage, err = template.ParseFiles("templates/repo.gmi")
	if err != nil {
		return err
	}
	repoPublicPage, err = template.ParseFiles("templates/public_repo.gmi")
	if err != nil {
		return err
	}
	groupListPage, err = template.ParseFiles("templates/group_list.gmi")
	if err != nil {
		return err
	}
	groupMembersPage, err = template.ParseFiles("templates/group.gmi")
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

func showIndex(c gig.Context) (error) {
	_, connected := db.GetUser(c.CertHash())
	data := struct {
		Title string
		Registration bool
		Connected bool
	}{
		Title: config.Cfg.Title,
		Registration: config.Cfg.Users.Registration,
		Connected: connected,
	}
	var b bytes.Buffer
	err := mainPage.Execute(&b, data)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure, err.Error())
	}
	return c.Gemini(b.String())
}

func showAccount(c gig.Context) (error) {
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
	data := struct {
		Username string
		Description string
		Repositories []string
		RepositoriesAccess []string
	}{
		Username: user.Name,
		Description: user.Description,
		Repositories: repoNames,
		RepositoriesAccess: nil,
	}
	var b bytes.Buffer
	err = accountPage.Execute(&b, data)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure, err.Error())
	}
	return c.Gemini(b.String())
}

func showGroups(c gig.Context) (error) {
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
	var b bytes.Buffer
	err = groupListPage.Execute(&b, data)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure, err.Error())
	}
	return c.Gemini(b.String())
}

func showMembers(c gig.Context) (error) {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid username")
	}
	group := c.Param("group")
	owner, err := user.IsInGroup(group)
	if err != nil {
		log.Println(err.Error())
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
	data := struct {
		Members []db.Member
		Owner bool
		Group string
		Description string
	}{
		Group: group,
		Owner: owner,
		Members: members,
		Description: desc,
	}
	var b bytes.Buffer
	err = groupMembersPage.Execute(&b, data)
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure, err.Error())
	}
	return c.Gemini(b.String())
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

func getPage(param string) (int) {
	switch param {
	case "":
		return pageLog
	case "files":
		return pageFiles
	case "refs":
		return pageRefs
	case "readme":
		return pageReadme
	case "license":
		return pageLicense
	}
	return -1;
}

func showRepo(c gig.Context, param string, owner bool) (error) {
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
	if !public && !owner {
		return c.NoContent(gig.StatusBadRequest,
				   "Private repository")
	}

	isEmpty := false
	content := ""
	commits := []commit{}
	files := []file{}
	branches := []branch{}
	tags := []branch{}
	page := getPage(param)
	switch page {
	case pageLog:
		ret, err := repo.GetCommits(name, author)
		if err != nil {
			log.Println(err.Error())
			return c.NoContent(gig.StatusBadRequest,
					   "Corrupted repository")
		}
		if ret == nil {
			isEmpty = true
			break
		} 
		err = ret.ForEach(func(c *object.Commit) error {
			info := c.Hash.String() + ", by " + c.Author.Name +
				" on " +
				c.Author.When.Format("2006-01-02 15:04:05")
			commits = append(commits,
					 commit{Info: info,
					 	Message: c.Message})
			return nil
		})
	case pageFiles:
		ret, err := repo.GetFiles(name, author)
		if err != nil {
			log.Println(err.Error())
			return c.NoContent(gig.StatusBadRequest,
					   "Corrupted repository")
		}
		if ret == nil {
			isEmpty = true
			break
		} 
		err = ret.ForEach(func(f *object.File) error {
			info := f.Mode.String() + " " + f.Name +
				" " + strconv.Itoa(int(f.Size))
			files = append(files,
					file{Info: info,
					Hash: f.Blob.Hash.String()})
			return nil
		})
	case pageRefs:
		refs, err := repo.GetRefs(name, author)
		if err != nil {
			log.Println(err)
			return c.NoContent(gig.StatusBadRequest,
					   "Corrupted repository")
		}
		if refs == nil {
			isEmpty = true
			break
		}
		err = refs.ForEach(func(c *plumbing.Reference) error {
			if c.Type().String() != "hash-reference" ||
			   c.Name().IsRemote() {
				return nil
			}
			var b branch
			b.Name = c.Name().String()
			name = name[strings.LastIndex(name, "/") + 1:]
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
	case pageLicense:
		content, err = showRepoFile(author, name, "LICENSE")
		if err != nil {
			log.Println(err.Error())
			return c.NoContent(gig.StatusBadRequest,
					   "Not license found")
		}
	case pageReadme:
		content, err = showRepoFile(author, name, "README.gmi")
		if err != nil {
			content, err = showRepoFile(author, name, "README")
		}
		if err != nil {
			log.Println(err.Error())
			return c.NoContent(gig.StatusBadRequest,
					   "Not readme found")
		}
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
		Log bool
		Files bool
		Refs bool
		Readme bool
		License bool
		Empty bool
		Commits []commit
		Tags []branch
		Branches []branch
		FileList []file
		Content string
	}{
		Public: public,
		Protocol: protocol,
		Domain: config.Cfg.Git.Domain,
		User: author,
		Description: desc,
		Repo: name,
		HasReadme: hasFile(name, author, "README.gmi") ||
			   hasFile(name, author, "README"),
		HasLicense: hasFile(name, author, "LICENSE"),
		Log: page == pageLog,
		Readme: page == pageReadme,
		License: page == pageLicense,
		Files: page == pageFiles,
		Refs: page == pageRefs,
		Commits: commits,
		Tags: tags,
		Branches: branches,
		FileList: files,
		Empty: isEmpty,
		Content: content,
	}
	var b bytes.Buffer
	if owner {
		err = repoPage.Execute(&b, data)
	} else {
		err = repoPublicPage.Execute(&b, data)
	}
	if err != nil {
		log.Println(err.Error())
		return c.NoContent(gig.StatusTemporaryFailure, err.Error())
	}
	return c.Gemini(b.String())
}
