package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"golang.org/x/crypto/ssh/terminal"

	"gemigit/auth"
	"gemigit/config"
	"gemigit/db"
	"gemigit/httpgit"
	"gemigit/repo"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pitr/gig"
)

func main() {

	if err := config.LoadConfig(); err != nil {
		log.Fatalln(err.Error())
	}

	if len(os.Args) > 1 {
		switch (os.Args[1]) {
		case "chpasswd":
			if len(os.Args) < 3 {
				fmt.Println(os.Args[0] + " chpasswd <username>")
				return
			}
			fmt.Print("New Password : ")
			password, err := terminal.ReadPassword(0)
			fmt.Print("\n")
			if err != nil {
				log.Fatalln(err.Error())
			}
			err = db.Init(config.Cfg.Gemigit.Database)
			if err != nil {
				log.Fatalln(err.Error())
			}
			defer db.Close()
			if err := db.ChangePassword(os.Args[2], string(password));
			   err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(os.Args[2] + "'s password changed")
			return
		case "register":
			if len(os.Args) < 3 {
				fmt.Println(os.Args[0] + " register <username>")
				return
			}
			fmt.Print("Password : ")
			password, err := terminal.ReadPassword(0)
			fmt.Print("\n")
			if err != nil {
				log.Fatalln(err.Error())
			}
			err = db.Init(config.Cfg.Gemigit.Database)
			if err != nil {
				log.Fatalln(err.Error())
			}
			defer db.Close()
			if err := db.Register(os.Args[2], string(password)); err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("User " + os.Args[2] + " created")
			return
		case "rmuser":
			if len(os.Args) < 3 {
				fmt.Println(os.Args[0] + " rmuser <username>")
				return
			}
			err := db.Init(config.Cfg.Gemigit.Database)
			if err != nil {
				log.Fatalln(err.Error())
			}
			defer db.Close()
			err = db.DeleteUser(os.Args[2])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("User " + os.Args[2] + " deleted successfully")
			return
		}
		fmt.Println("usage: " + os.Args[0] + " [command]")
		fmt.Println("commands :")
		fmt.Println("\tchpasswd <username>")
		fmt.Println("\tregister <username>")
		fmt.Println("\trmuser <username>")
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := db.Init(config.Cfg.Gemigit.Database)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer db.Close()
	if err := repo.Init("repos"); err != nil {
		log.Fatalln(err.Error())
	}

	go httpgit.Listen("repos/", config.Cfg.Gemigit.Port)
	go auth.Decrease()

	gig.DefaultLoggerConfig.Format = "${time_rfc3339} - ${remote_ip} | " +
					 "Path=${path}, Status=${status}, " +
					 "Latency=${latency}\n"
	g := gig.Default()
	g.Use(gig.Recover())

	secure := g.Group("/account", gig.PassAuth(
	func(sig string, c gig.Context) (string, error) {
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
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
			}
			ret :=
			"=>/ Main page\n\n" +
			"# Account : " + user.Name + "\n"
			err := user.UpdateDescription()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if user.Description != "" {
				ret += user.Description + "\n\n"
			} else {
				ret += "\n"
			}
			ret +=
			"=>/account/addrepo Create a new repository\n" +
			"=>/account/chdesc Change your account description\n" +
			"=>/account/chpasswd Change your password\n" +
			"=>/account/disconnect Disconnect\n" +
			"\n## Repositories list\n\n"

			repos, err := user.GetRepos(false)
			if err != nil {
				ret += "Failed to load user's repositories\n"
				log.Println(err)
				return c.Gemini(ret)
			}
			for _, repo := range repos {
				ret += "=>/account/repo/" + repo.Name + 
					" " + repo.Name + "\n"
			}

			return c.Gemini(ret)
		})

		secure.Handle("/repo/:repo/files", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
			}
			query, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if query == "" {
				return repoRequest(c, "files", true)
			}
			repofile, err := repo.GetFile(c.Param("repo"), user.Name, query)
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			contents, err := repofile.Contents()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.Gemini(contents)
		})

		secure.Handle("/repo/:repo/files/:blob", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
			}
			content, err := repo.GetPrivateFile(c.Param("repo"),
							    user.Name,
							    c.Param("blob"),
							    c.CertHash())
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   err.Error())
			}
			lines := strings.Split(content, "\n")
			file := ""
			for i, line := range lines {
				file += strconv.Itoa(i) + " \t" + line + "\n"
			}
			return c.Gemini(file)
		})

		secure.Handle("/repo/:repo/refs", func(c gig.Context) error {
			return repoRequest(c, "refs", true)
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
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
			}
			repofile, err := repo.GetFile(c.Param("repo"),
						      user.Name,
						      c.Param("*"))
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
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
			}
			if err := user.TogglePublic(c.Param("repo"), c.CertHash());
			   err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary,
					   "/account/repo/" + c.Param("repo"))
		})

		secure.Handle("/repo/:repo/chname", func(c gig.Context) error {
			newname, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid input received")
			}
			if newname == "" {
				return c.NoContent(gig.StatusInput, "New repository name")
			}
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest, "Invalid username")
			}
			if err := user.ChangeRepoName(c.Param("repo"), newname, c.CertHash());
			   err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if err := repo.ChangeRepoDir(c.Param("repo"), user.Name, newname);
			   err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary,
					   "/account/repo/" + newname)
		})

		secure.Handle("/repo/:repo/chdesc", func(c gig.Context) error {
			newdesc, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid input received")
			}
			if newdesc == "" {
				return c.NoContent(gig.StatusInput,
						   "New repository description")
			}
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
			}
			if err := user.ChangeRepoDesc(c.Param("repo"), newdesc);
			   err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary,
					   "/account/repo/" + c.Param("repo"))
		})

		secure.Handle("/chdesc", func(c gig.Context) error {
			newdesc, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid input received")
			}
			if newdesc == "" {
				return c.NoContent(gig.StatusInput,
						   "New account description")
			}
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
			}
			if err := user.ChangeDescription(newdesc, c.CertHash());
			   err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary, "/account")
		})

		secure.Handle("/addrepo", func(c gig.Context) error {

			name, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			if name == "" {
				return c.NoContent(gig.StatusInput, "Repository name")
			}
			user, b := db.GetUser(c.CertHash())
			if !b {
				return c.NoContent(gig.StatusBadRequest,
						   "Cannot find username")
			}
			if err := user.CreateRepo(name, c.CertHash());
			   err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   err.Error())
			}
			if err := repo.InitRepo(name, user.Name); err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary,
					   "/account/repo/" + name)

		})

		secure.Handle("/repo/:repo/delrepo", func(c gig.Context) error {

			name, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid input received")
			}
			if name != "" {
				return c.NoContent(gig.StatusInput,
						   "Type the repository name")
			}
			if name != c.Param("repo") {
				return c.NoContent(gig.StatusRedirectTemporary,
						   "/account/repo/" + 
						   c.Param("repo"))
			}
			user, b := db.GetUser(c.CertHash())
			if !b {
				return c.NoContent(gig.StatusBadRequest,
						   "Cannot find username")
			}
			if err := user.DeleteRepo(name, c.CertHash());
			   err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   err.Error())
			}
			if err := repo.RemoveRepo(name, user.Name);
			   err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary,
					   "/account")
		})

		secure.Handle("/chpasswd", func(c gig.Context) error {
			passwd, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid input received")
			}
			if passwd == "" {
				return c.NoContent(gig.StatusSensitiveInput,
						   "New password")
			}
			user, b := db.GetUser(c.CertHash())
			if !b {
				return c.NoContent(gig.StatusBadRequest,
						   "Cannot find username")
			}
			err = user.ChangePassword(passwd, c.CertHash())
			if err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.NoContent(gig.StatusRedirectTemporary,
					   "/account")
		})

		secure.Handle("/disconnect", func(c gig.Context) error {
			user, exist := db.GetUser(c.CertHash())
			if !exist {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid username")
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
				return c.NoContent(gig.StatusTemporaryFailure,
						   "Internal error, "+err.Error())
			}
			for _, repo := range repos {
				ret += "=> /repo/" + repo.Username +
					"/" + repo.Name + " " + 
					repo.Name + " by " + repo.Username + "\n"
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
				return c.NoContent(gig.StatusTemporaryFailure,
						   "Invalid account, " + err.Error())
			}
			for _, repo := range repos {
				ret += "=> /repo/" + repo.Username +
					"/" + repo.Name + " " + repo.Name + "\n"
			}
			return c.Gemini(ret)
		})

		public.Handle("/:user/:repo/files", func(c gig.Context) error {
			return repoRequest(c, "files", false)
		})

		public.Handle("/:user/:repo/files/:blob", func(c gig.Context) error {
			content, err := repo.GetPublicFile(c.Param("repo"), 
							   c.Param("user"),
							   c.Param("blob"))
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

		public.Handle("/:user/:repo/refs", func(c gig.Context) error {
			return repoRequest(c, "refs", false)
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
			repofile, err := repo.GetFile(c.Param("repo"),
						      c.Param("user"),
						      c.Param("*"))
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

	g.PassAuthLoginHandle("/login",
	func(user, pass, sig string, c gig.Context) (string, error) {
		err := auth.Connect(user, pass, sig, c.IP())
		if err != nil {
			return "", err
		}
		return "/account", nil
	})

	if config.Cfg.Gemigit.AllowRegistration {
		g.Handle("/register",
		func(c gig.Context) error {
			cert := c.Certificate()
			if cert == nil {
				return c.NoContent(gig.StatusClientCertificateRequired,
						   "Certificate required")
			}

			name, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid name received")
			}
			if name != "" {
				return c.NoContent(gig.StatusRedirectPermanent,
						   "/register/"+name)
			}

			return c.NoContent(gig.StatusInput, "Username")
		})

		g.Handle("/register/:name", func(c gig.Context) error {
			cert := c.Certificate()
			if cert == nil {
				return c.NoContent(gig.StatusClientCertificateRequired,
						   "Certificate required")
			}

			password, err := c.QueryString()
			if err != nil {
				return c.NoContent(gig.StatusBadRequest,
						   "Invalid password received")
			}
			if password != "" {
				return c.NoContent(gig.StatusSensitiveInput, "Password")
			}
			if err = db.Register(c.Param("name"), password); err != nil {
				return c.NoContent(gig.StatusBadRequest, err.Error())
			}
			return c.Gemini("# Your registration was completed successfully\n" +
					"=> /login Login now")
		})
	}

	g.Handle("/", func(c gig.Context) error {
		_, connected := db.GetUser(c.CertHash())
		ret := ""
		if !connected {
			ret = "# " + config.Cfg.Gemigit.Name + "\n\n"
			ret += "=> /login Login\n"
			if config.Cfg.Gemigit.AllowRegistration {
				ret += "=> /register Register\n"
			}
		} else {
			ret = "# " + config.Cfg.Gemigit.Name + "\n\n" +
			      "=> /account Account page\n"
		}
		ret += "=> /repo Public repositories"
		return c.Gemini(ret)
	})

	err = g.Run("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err.Error())
	}
}
