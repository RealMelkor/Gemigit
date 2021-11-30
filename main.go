package main

import (
	"log"
	"strconv"
	"strings"

	"gemigit/db"
	"gemigit/httpgit"
	"gemigit/repo"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pitr/gig"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := loadConfig(); err != nil {
		log.Fatalln(err.Error())
	}

	err := db.Init(cfg.Gemigit.Database)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer db.Close()
	if err := repo.Init("repos"); err != nil {
		log.Fatalln(err.Error())
	}

	go httpgit.Listen("repos/", cfg.Gemigit.Port)

	g := gig.Default()
	g.Use(gig.Recover())

	secure := g.Group("/account", gig.PassAuth(func(sig string, c gig.Context) (string, error) {
		_, b := db.GetUsername(sig)
		if !b {
			return "/login", nil
		}
		return "", nil
	}))
	{

		secure.Handle("/main", func(c gig.Context) error {
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			ret := "=>/ Main page\n\n"
			ret += "# Account : " + username + "\n\n"
			ret += "=>/account/addrepo Create a new repository\n"
			ret += "=>/account/delrepo Delete a new repository\n"
			ret += "\n## Repositories list\n\n"

			repos, err := db.GetRepoFromUser(username, false)
			if err != nil {
				ret += "Failed to load user's repositories\n"
			} else {
				for _, repo := range repos {
					ret += "=>/account/repo/" + repo.Name + " " + repo.Name + "\n"
				}
			}

			return c.Gemini(ret)
		})

		secure.Handle("/repo/:repo/:blob", func(c gig.Context) error {
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			content, err := repo.GetPrivateFile(c.Param("repo"), username, c.Param("blob"), c.CertHash())
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			lines := strings.Split(content, "\n")
			file := ""
			for i, line := range lines {
				file += strconv.Itoa(i) + " \t" + line + "\n"
			}
			return c.Gemini(file)
		})

		secure.Handle("/repo/:repo", func(c gig.Context) error {
			ret := "=>/account/main Go back\n\n"
			ret += "# " + c.Param("repo") + "\n\n"
			ret += "=>/account/repo/" + c.Param("repo") + "/togglepublic Make the repository "
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			b, err := db.IsRepoPublic(c.Param("repo"), username)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if b {
				ret += "private\n"
			} else {
				ret += "public\n"
			}

			commits, err := repo.GetCommits(c.Param("repo"), username)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			ret += "\n## Commits\n\n"
			if commits != nil {
				err = commits.ForEach(func(c *object.Commit) error {
					ret += "* " + c.Hash.String() + ", by " + c.Author.Name + " on " + c.Author.When.Format("2006-01-02 15:04:05") + "\n"
					ret += "> " + c.Message + "\n"
					return nil
				})
				if err != nil {
					return c.NoContent(gig.StatusBadRequest, err.Error())
				}
			} else {
				ret += "Empty repository\n"
			}
			ret += "\n## Files\n\n"
			files, err := repo.GetFiles(c.Param("repo"), username)
			if err != nil {
				log.Println(err.Error())
			}
			if files != nil {
				err = files.ForEach(func(f *object.File) error {
					ret += "=> " + c.Param("repo") + "/" + f.Blob.Hash.String() + " " + f.Mode.String() + " " + f.Name + " " + strconv.Itoa(int(f.Size)) + "\n"
					return nil
				})
				if err != nil {
					return c.NoContent(gig.StatusBadRequest, err.Error())
				}
			} else {
				ret += "Empty repository\n"
			}

			return c.Gemini(ret)
		})

		secure.Handle("/repo/:repo/togglepublic", func(c gig.Context) error {
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := db.TogglePublic(c.Param("repo"), username); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary, "/account/repo/"+c.Param("repo"))
		})

		secure.Handle("/repo", func(c gig.Context) error {
			return c.Gemini("# " + c.Param("name") + "'s user page")
		})

		secure.Handle("/addrepo", func(c gig.Context) error {

			name, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid input received")
			}
			if name != "" {
				username, b := db.GetUsername(c.CertHash())
				if b {
					db.CreateRepo(name, username)
					repo.InitRepo(name, username)
					return c.NoContent(gig.StatusRedirectTemporary, "/account/repo/"+name)
				}
				return c.NoContent(gig.StatusBadRequest, "Cannot find username")
			}

			return c.NoContent(gig.StatusInput, "Repository name")
		})

		secure.Handle("/delrepo", func(c gig.Context) error {

			name, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid input received")
			}
			if name != "" {
				username, b := db.GetUsername(c.CertHash())
				if b {
					if err := db.DeleteRepo(name, username, c.CertHash()); err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
					if err := repo.RemoveRepo(name, username); err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
					return c.NoContent(gig.StatusRedirectTemporary, "/account/main")
				}
				return c.NoContent(gig.StatusBadRequest, "Cannot find username")
			}

			return c.NoContent(gig.StatusInput, "Repository name")
		})
	}

	public := g.Group("/repo")
	{
		g.Handle("/repo", func(c gig.Context) error {
			ret := "=>/ Go back\n\n"
			ret += "# Public repositories\n"
			repos, err := db.GetPublicRepo()
			if err != nil {
				return c.NoContent(gig.StatusTemporaryFailure, "Internal error, "+err.Error())
			}
			for _, repo := range repos {
				ret += "=> /repo/" + repo.Username + "/" + repo.Name + " " + repo.Name + " by " + repo.Username + "\n"
			}
			return c.Gemini(ret)
		})

		public.Handle("/:user", func(c gig.Context) error {
			ret := "=>/repo Go back\n\n"
			ret += "# " + c.Param("user") + "'s repositories\n"
			repos, err := db.GetRepoFromUser(c.Param("user"), true)
			if err != nil {
				return c.NoContent(gig.StatusTemporaryFailure, "Invalid account, "+err.Error())
			}
			for _, repo := range repos {
				ret += "=> /repo/" + repo.Username + "/" + repo.Name + " " + repo.Name + "\n"
			}
			return c.Gemini(ret)
		})

		public.Handle("/:user/:repo", func(c gig.Context) error {
			ret := "=>/repo Go back\n\n"
			ret += "# " + c.Param("repo") + " by " + c.Param("user") + "\n"
			ret += "=>/repo/" + c.Param("user") + " View account" + "\n\n"
			ret += "> git clone "
			if cfg.Gemigit.Https {
				ret += "https://"
			} else {
				ret += "http://"
			}
			ret += cfg.Gemigit.Domain + "/" + c.Param("user") + "/" + c.Param("repo") + "\n"
			public, err := db.IsRepoPublic(c.Param("repo"), c.Param("user"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if !public {
				return c.NoContent(gig.StatusBadRequest, "repository not found")
			}
			commits, err := repo.GetCommits(c.Param("repo"), c.Param("user"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			ret += "\n## Commits\n\n"
			if commits != nil {
				err = commits.ForEach(func(c *object.Commit) error {
					ret += "* " + c.Hash.String() + ", by " + c.Author.Name + " on " + c.Author.When.Format("2006-01-02 15:04:05") + "\n"
					ret += "> " + c.Message + "\n"
					return nil
				})
				if err != nil {
					return c.NoContent(gig.StatusBadRequest, err.Error())
				}
			} else {
				ret += "Empty repository\n"
			}
			ret += "\n## Files\n\n"
			files, err := repo.GetFiles(c.Param("repo"), c.Param("user"))
			if err != nil {
				log.Println(err.Error())
			}
			if files != nil {
				err = files.ForEach(func(f *object.File) error {
					ret += "=> " + c.Param("repo") + "/" + f.Blob.Hash.String() + " " + f.Mode.String() + " " + f.Name + " " + strconv.Itoa(int(f.Size)) + "\n"
					return nil
				})
				if err != nil {
					return c.NoContent(gig.StatusBadRequest, err.Error())
				}
			} else {
				ret += "Empty repository\n"
			}
			return c.Gemini(ret)
		})

		public.Handle("/:user/:repo/:blob", func(c gig.Context) error {
			content, err := repo.GetPublicFile(c.Param("repo"), c.Param("user"), c.Param("blob"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			lines := strings.Split(content, "\n")
			file := ""
			for i, line := range lines {
				file += strconv.Itoa(i) + " \t" + line + "\n"
			}
			return c.Gemini(file)
		})
	}

	g.Handle("/error/:error", func(c gig.Context) error {
		content := ""
		switch c.Param("error") {
		case "login":
			content += "# Failed to login"
		case "internal":
			content += "# Internal error"
		default:
			content += "# Unknown error"
		}
		content += "\n\n=> / Go back to the main page"
		return c.Gemini(content)
	})

	g.PassAuthLoginHandle("/login", func(user, pass, sig string, c gig.Context) (string, error) {
		success, err := db.Login(user, pass, sig)
		if err != nil {
			return "/error/internal", err
		}
		if !success {
			return "/error/login", nil
		}
		return "/account/main", nil
	})

	if cfg.Gemigit.AllowRegistration {
		g.Handle("/register", func(c gig.Context) error {
			cert := c.Certificate()
			if cert == nil {
				return c.NoContent(gig.StatusClientCertificateRequired, "Certificate required")
			}

			name, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid name received")
			}
			if name != "" {
				return c.NoContent(gig.StatusRedirectPermanent, "/register/"+name)
			}

			return c.NoContent(gig.StatusInput, "Username")
		})

		g.Handle("/register/:name", func(c gig.Context) error {
			cert := c.Certificate()
			if cert == nil {
				return c.NoContent(gig.StatusClientCertificateRequired, "Certificate required")
			}

			password, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid password received")
			}
			if password != "" {
				if err = db.Register(c.Param("name"), password); err != nil {
					return c.Gemini("# Registration failure\n" + err.Error() + "\n\n=> /register Retry?\n\n=> / Go back to the main page")
				}
				return c.Gemini("# Your registration was completed successfully\n=> /login Login now")
			}

			return c.NoContent(gig.StatusSensitiveInput, "Password")
		})
	}

	g.Handle("/", func(c gig.Context) error {
		_, connected := db.GetUsername(c.CertHash())
		ret := ""
		if !connected {
			ret = "# " + cfg.Gemigit.Name + "\n=> /login Login\n"
			if cfg.Gemigit.AllowRegistration {
				ret += "=> /register Register\n"
			}
		} else {
			ret = "# " + cfg.Gemigit.Name + "\n=> /account/main Account page\n"
		}
		ret += "=> /repo Public repositories"
		return c.Gemini(ret)
	})

	err = g.Run("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err.Error())
	}
}
