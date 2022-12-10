package repo

import (
	"errors"
	"gemigit/config"
	"gemigit/db"
	"io"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var rootPath string
type repo struct {
	memory storage.Storer
	repository *git.Repository
	update time.Time
}
var repositories = make(map[string]repo)

func Init(path string) error {
	if config.Cfg.Git.Remote.Enabled {
		return nil
	}
	rootPath = path
	return os.MkdirAll(path, 0700)
}

func InitRepo(name string, username string) error {
	if config.Cfg.Git.Remote.Enabled {
		err := request("api/" + config.Cfg.Git.Remote.Key + "/init/" +
			       username + "/" + name)
		return err
	}
	_, err := git.PlainInit(rootPath+"/"+username+"/"+name, true)
	return err
}

func RemoveRepo(name string, username string) error {
	if config.Cfg.Git.Remote.Enabled {
		err := request("api/" + config.Cfg.Git.Remote.Key + "/rm/" +
			       username + "/" + name)
		return err
	}
	return os.RemoveAll(rootPath + "/" + username + "/" + name)
}

func getRepo(name string, username string) (*git.Repository, error) {
	var repository *git.Repository
	var err error
	url := username + "/" + name
	if !config.Cfg.Git.Remote.Enabled {
		repository, err = git.PlainOpen(rootPath + "/" + url)
		return repository, err
	}
	var exist bool
	r, exist := repositories[url]
	if !exist {
		r = repo{}
	}
	if exist && time.Now().Sub(r.update).Seconds() < 15 {
		return r.repository, nil
	}
	r.update = time.Now()
	repositories[url] = r
	if r.repository == nil {
		r.memory = memory.NewStorage()
		r.repository, err = git.Clone(r.memory, nil,
		&git.CloneOptions {
			URL: config.Cfg.Git.Remote.Url + "/" + url,
			Auth: &http.BasicAuth {
				Username: "root#",
				Password: config.Cfg.Git.Remote.Key,
			},
		})
		if err != nil {
			r.memory = nil
			r.repository = nil
		}
		repositories[url] = r
		return r.repository, err
	}
	err = r.repository.Fetch(&git.FetchOptions {
		Auth: &http.BasicAuth {
			Username: "root#",
			Password: config.Cfg.Git.Remote.Key,
		},
	})
	if err != nil {
		return nil, err
	}
	repositories[url] = r
	return r.repository, nil
}

func GetCommit(name string, username string,
	       hash plumbing.Hash) (*object.Commit, error) {
	repo, err := getRepo(name, username)
	if err != nil {
		return nil, err
	}
	obj, err := repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func GetCommits(name string, username string) (object.CommitIter, error) {
	repo, err := getRepo(name, username)
	if repo == nil || err != nil {
		return nil, err
	}
	ref, err := repo.Head()
	if err != nil {
		return nil, nil // Empty repo
	}
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	return cIter, nil
}

func GetRefs(name string, username string) (storer.ReferenceIter, error) {
	repo, err := getRepo(name, username)
	if repo == nil || err != nil {
		return nil, err
	}
	_, err = repo.Head()
	if err != nil {
		return nil, nil // Empty repo
	}
	refs, err := repo.References()
	if err != nil {
		return nil, err
	}
	return refs, nil
}

func getTree(name string, username string) (*object.Tree, error) {
	repo, err := getRepo(name, username)
	if repo == nil || err != nil {
		return nil, err
	}
	ref, err := repo.Head()
	if err != nil {
		return nil, nil // Empty repo
	}
	last, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	tree, err := repo.TreeObject(last.TreeHash)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func GetFiles(name string, username string) (*object.FileIter, error) {
	tree, err := getTree(name, username)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return nil, nil
	}
	return tree.Files(), nil
}

func GetFile(name string, username string, file string) (*object.File, error) {
	tree, err := getTree(name, username)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return nil, nil
	}
	out, err := tree.File(file)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetPublicFile(name string, username string, hash string) (string, error) {
	public, err := db.IsRepoPublic(name, username)
	if err != nil {
		return "", err
	}
	if !public {
		return "", errors.New("repository is private")
	}
	repo, err := git.PlainOpen(rootPath + "/" + username + "/" + name)
	if err != nil {
		return "", err
	}
	file, err := repo.BlobObject(plumbing.NewHash(hash))
	if err != nil {
		return "", err
	}
	reader, err := file.Reader()
	if err != nil {
		return "", err
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func GetPrivateFile(name string, username string,
		    hash string, sig string) (string, error) {
	user, b := db.GetUser(sig)
	if !b || username != user.Name {
		return "", errors.New("invalid signature")
	}
	repo, err := git.PlainOpen(rootPath + "/" + username + "/" + name)
	if err != nil {
		return "", err
	}
	file, err := repo.BlobObject(plumbing.NewHash(hash))
	if err != nil {
		return "", err
	}
	reader, err := file.Reader()
	if err != nil {
		return "", err
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func ChangeRepoDir(name string, username string, newname string) error {
	if config.Cfg.Git.Remote.Enabled {
		err := request("api/" + config.Cfg.Git.Remote.Key + "/mv/" +
			       username + "/" + name + "/" + newname)
		return err
	}
	return os.Rename(rootPath + "/" + username + "/" + name,
			 rootPath + "/" + username + "/" + newname)
}
