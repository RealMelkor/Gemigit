package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"gemigit/db"
	"gemigit/httpgit"
	"gemigit/repo"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pitr/gig"
)

func getHttpAddress(user string, repo string) string {
	ret := "git clone "
	if cfg.Gemigit.Https {
		ret += "https://"
	} else {
		ret += "http://"
	}
	ret += cfg.Gemigit.Domain + "/" + user + "/" + repo + "\n"
	return ret
}

func showRepoHeader(user string, reponame string, owner bool) (string, error) {
	ret := ""
	if owner {
		ret += "=>/account/main Go back\n\n"
	} else {
		ret += "=>/repo Go back\n\n"
	}
	ret += "# " + reponame + "\n"
	desc, err := db.GetRepoDesc(reponame, user)
	if err != nil {
		return "", err
	}
	if desc != "" {
		ret += "> " + desc + "\n"
	}
	ret += "> " + getHttpAddress(user, reponame)
	if owner {
		ret += "\n"
		ret += "=>/account/repo/" + reponame + "/chname Change repository name\n"
		ret += "=>/account/repo/" + reponame + "/chdesc Change repository description\n"
		ret += "=>/account/repo/" + reponame + "/togglepublic Make the repository "
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
			ret += "=>/repo/" + user + "/" + reponame + "/license License\n"
		}
		file, err = repo.GetFile(reponame, user, "README")
		if file != nil && err == nil {
			ret += "=>/repo/" + user + "/" + reponame + "/readme Readme\n"
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
			ret += reponame + "/" + f.Blob.Hash.String() + " " + f.Mode.String() + " " + f.Name + " " + strconv.Itoa(int(f.Size)) + "\n"
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
			ret += "* " + c.Hash.String() + ", by " + c.Author.Name + " on " + c.Author.When.Format("2006-01-02 15:04:05") + "\n"
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

func main() {

	if err := loadConfig(); err != nil {
		log.Fatalln(err.Error())
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "chpasswd" {
			if len(os.Args) < 4 {
				fmt.Println(os.Args[0] + " chpasswd <username> <new password>")
				return
			}
			err := db.Init(cfg.Gemigit.Database)
			if err != nil {
				log.Fatalln(err.Error())
			}
			defer db.Close()
			if err := db.ChangePassword(os.Args[2], os.Args[3]); err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(os.Args[2] + "'s password changed")
			return
		}
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := db.Init(cfg.Gemigit.Database)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer db.Close()
	if err := repo.Init("repos"); err != nil {
		log.Fatalln(err.Error())
	}

	go httpgit.Listen("repos/", cfg.Gemigit.Port)

	gig.DefaultLoggerConfig.Format = "${time_rfc3339} - ${remote_ip} | Path=${path}, Status=${status}, Latency=${latency}\n"
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
			ret += "=>/account/chpasswd Change your password\n"
			ret += "=>/account/disconnect Disconnect\n"
			ret += "\n## Repositories list\n\n"

			repos, err := db.GetRepoFromUser(username, false)
			if err != nil {
				ret += "Failed to load user's repositories\n"
				log.Println(err)
			} else {
				for _, repo := range repos {
					ret += "=>/account/repo/" + repo.Name + " " + repo.Name + "\n"
				}
			}

			return c.Gemini(ret)
		})

		secure.Handle("/repo/:repo/files", func(c gig.Context) error {
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			ret, err := showRepoHeader(username, c.Param("repo"), true)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showRepoFiles(username, c.Param("repo"), true)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			ret += out
			return c.Gemini(ret)
		})

		secure.Handle("/repo/:repo/license", func(c gig.Context) error {
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			ret, err := showRepoHeader(username, c.Param("repo"), true)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showLicense(username, c.Param("repo"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			ret += out
			return c.Gemini(ret)
		})

		secure.Handle("/repo/:repo/readme", func(c gig.Context) error {
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			ret, err := showRepoHeader(username, c.Param("repo"), true)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showReadme(username, c.Param("repo"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			ret += out
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
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			ret, err := showRepoHeader(username, c.Param("repo"), true)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showRepoCommits(username, c.Param("repo"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			ret += out
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

		secure.Handle("/repo/:repo/chname", func(c gig.Context) error {
			newname, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid input received")
			}
			if newname == "" {
				return c.NoContent(gig.StatusInput, "New repository name")
			}
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := db.ChangeRepoName(c.Param("repo"), username, newname); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if err := repo.ChangeRepoDir(c.Param("repo"), username, newname); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary, "/account/repo/"+newname)
		})

		secure.Handle("/repo/:repo/chdesc", func(c gig.Context) error {
			newdesc, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid input received")
			}
			if newdesc == "" {
				return c.NoContent(gig.StatusInput, "New repository description")
			}
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := db.ChangeRepoDesc(c.Param("repo"), username, newdesc); err != nil {
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

		secure.Handle("/chpasswd", func(c gig.Context) error {
			passwd, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid input received")
			}
			if passwd != "" {

				username, b := db.GetUsername(c.CertHash())
				if b {
					err := db.ChangePassword(username, passwd)
					if err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
					return c.NoContent(gig.StatusRedirectTemporary, "/account/main")
				}
				return c.NoContent(gig.StatusBadRequest, "Cannot find username")
			}
			return c.NoContent(gig.StatusSensitiveInput, "New password")
		})

		secure.Handle("/disconnect", func(c gig.Context) error {
			username, exist := db.GetUsername(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := db.Disconnect(username, c.CertHash()); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary, "/")
		})
	}

	public := g.Group("/repo")
	{
		g.Handle("/repo", func(c gig.Context) error {
			ret := "=>/ Go back\n\n"
			ret += "# Public repositories\n\n"
			repos, err := db.GetPublicRepo()
			if err != nil {
				log.Println(err.Error())
				return c.NoContent(gig.StatusTemporaryFailure, "Internal error, "+err.Error())
			}
			for _, repo := range repos {
				ret += "=> /repo/" + repo.Username + "/" + repo.Name + " " + repo.Name + " by " + repo.Username + "\n"
				if repo.Description != "" {
					ret += "> " + repo.Description + "\n"
				}
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

		public.Handle("/:user/:repo/files", func(c gig.Context) error {
			ret, err := showRepoHeader(c.Param("user"), c.Param("repo"), false)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showRepoFiles(c.Param("user"), c.Param("repo"), false)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.Gemini(ret + out)
		})

		public.Handle("/:user/:repo/license", func(c gig.Context) error {
			ret, err := showRepoHeader(c.Param("user"), c.Param("repo"), false)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showLicense(c.Param("user"), c.Param("repo"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.Gemini(ret + out)
		})

		public.Handle("/:user/:repo/readme", func(c gig.Context) error {
			ret, err := showRepoHeader(c.Param("user"), c.Param("repo"), false)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showReadme(c.Param("user"), c.Param("repo"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.Gemini(ret + out)
		})

		public.Handle("/:user/:repo", func(c gig.Context) error {
			ret, err := showRepoHeader(c.Param("user"), c.Param("repo"), false)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			out, err := showRepoCommits(c.Param("user"), c.Param("repo"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.Gemini(ret + out)
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
			ret = "# " + cfg.Gemigit.Name + "\n\n"
			ret += "=> /login Login\n"
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
