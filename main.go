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

	"github.com/gabriel-vasile/mimetype"
	"github.com/pitr/gig"
)

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
		_, b := db.GetUser(sig)
		if !b {
			return "/login", nil
		}
		return "", nil
	}))
	{

		secure.Handle("", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			ret := "=>/ Main page\n\n"
			ret += "# Account : " + user.Name + "\n"
			if user.Description != "" {
				ret += user.Description + "\n\n"
			} else {
				ret += "\n"
			}
			ret += "=>/account/addrepo Create a new repository\n"
			ret += "=>/account/delrepo Delete a new repository\n"
			ret += "=>/account/chdesc Change your account description\n"
			ret += "=>/account/chpasswd Change your password\n"
			ret += "=>/account/disconnect Disconnect\n"
			ret += "\n## Repositories list\n\n"

			repos, err := user.GetRepos(false)
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
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			query, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if query != "" {
				repofile, err := repo.GetFile(c.Param("repo"), user.Name, query)
				if err != nil {
					return c.NoContent(gig.StatusBadRequest, err.Error())
				}
				contents, err := repofile.Contents()
				if err != nil {
					return c.NoContent(gig.StatusBadRequest, err.Error())
				}
				return c.Gemini(contents)
			}
			return repoRequest(c, "files", true)
		})

		secure.Handle("/repo/:repo/files/:blob", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			content, err := repo.GetPrivateFile(c.Param("repo"), user.Name, c.Param("blob"), c.CertHash())
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

		secure.Handle("/repo/:repo/license", func(c gig.Context) error {
			return repoRequest(c, "license", true)
		})

		secure.Handle("/repo/:repo/readme", func(c gig.Context) error {
			return repoRequest(c, "readme", true)
		})

		secure.Handle("/repo/:repo/*", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			repofile, err := repo.GetFile(c.Param("repo"), user.Name, c.Param("*"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			reader, err := repofile.Reader()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			buf, err := io.ReadAll(reader)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			mtype := mimetype.Detect(buf)
			return c.Blob(mtype.String(), buf)
		})

		secure.Handle("/repo/:repo", func(c gig.Context) error {
			return repoRequest(c, "", true)
		})

		secure.Handle("/repo/:repo/togglepublic", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := user.TogglePublic(c.Param("repo"), c.CertHash()); err != nil {
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
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := user.ChangeRepoName(c.Param("repo"), newname, c.CertHash()); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if err := repo.ChangeRepoDir(c.Param("repo"), user.Name, newname); err != nil {
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
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := user.ChangeRepoDesc(c.Param("repo"), newdesc); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary, "/account/repo/"+c.Param("repo"))
		})

		secure.Handle("/chdesc", func(c gig.Context) error {
			newdesc, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, "Invalid input received")
			}
			if newdesc == "" {
				return c.NoContent(gig.StatusInput, "New account description")
			}
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := user.ChangeDescription(newdesc, c.CertHash()); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary, "/account")
		})

		secure.Handle("/addrepo", func(c gig.Context) error {

			name, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if name != "" {
				user, b := db.GetUser(c.CertHash())
				if b {
					if err := user.CreateRepo(name); err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
					if err := repo.InitRepo(name, user.Name); err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
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
				user, b := db.GetUser(c.CertHash())
				if b {
					if err := user.DeleteRepo(name, c.CertHash()); err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
					if err := repo.RemoveRepo(name, user.Name); err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
					return c.NoContent(gig.StatusRedirectTemporary, "/account")
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

				user, b := db.GetUser(c.CertHash())
				if b {
					err := user.ChangePassword(passwd, c.CertHash())
					if err != nil {
						return c.NoContent(gig.StatusBadRequest, err.Error())
					}
					return c.NoContent(gig.StatusRedirectTemporary, "/account")
				}
				return c.NoContent(gig.StatusBadRequest, "Cannot find username")
			}
			return c.NoContent(gig.StatusSensitiveInput, "New password")
		})

		secure.Handle("/disconnect", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := user.Disconnect(c.CertHash()); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary, "/")
		})
	}

	public := g.Group("/repo")
	{
		public.Handle("", func(c gig.Context) error {
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
			ret := "=>/repo Go back\n\n# " + c.Param("user") + "\n\n"
			user, err := db.GetPublicUser(c.Param("user"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if user.Description != "" {
				ret += user.Description + "\n\n"
			}
			ret += "## Repositories\n"
			repos, err := user.GetRepos(true)
			if err != nil {
				return c.NoContent(gig.StatusTemporaryFailure, "Invalid account, "+err.Error())
			}
			for _, repo := range repos {
				ret += "=> /repo/" + repo.Username + "/" + repo.Name + " " + repo.Name + "\n"
			}
			return c.Gemini(ret)
		})

		public.Handle("/:user/:repo/files", func(c gig.Context) error {
			return repoRequest(c, "files", false)
		})

		public.Handle("/:user/:repo/files/:blob", func(c gig.Context) error {
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

		public.Handle("/:user/:repo/license", func(c gig.Context) error {
			return repoRequest(c, "license", false)
		})

		public.Handle("/:user/:repo/readme", func(c gig.Context) error {
			return repoRequest(c, "readme", false)
		})

		public.Handle("/:user/:repo", func(c gig.Context) error {
			return repoRequest(c, "", false)
		})

		public.Handle("/:user/:repo/*", func(c gig.Context) error {
			repofile, err := repo.GetFile(c.Param("repo"), c.Param("user"), c.Param("*"))
			log.Println(c.Param("*"))
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			reader, err := repofile.Reader()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			buf, err := io.ReadAll(reader)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			mtype := mimetype.Detect(buf)
			return c.Blob(mtype.String(), buf)
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
		return "/account", nil
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
		_, connected := db.GetUser(c.CertHash())
		ret := ""
		if !connected {
			ret = "# " + cfg.Gemigit.Name + "\n\n"
			ret += "=> /login Login\n"
			if cfg.Gemigit.AllowRegistration {
				ret += "=> /register Register\n"
			}
		} else {
			ret = "# " + cfg.Gemigit.Name + "\n=> /account Account page\n"
		}
		ret += "=> /repo Public repositories"
		return c.Gemini(ret)
	})

	err = g.Run("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err.Error())
	}
}
