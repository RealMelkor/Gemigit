package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/term"

	"gemigit/access"
	"gemigit/auth"
	"gemigit/config"
	"gemigit/db"
	"gemigit/httpgit"
	"gemigit/repo"
	"gemigit/gmi"
	"gemigit/csrf"

	"github.com/pitr/gig"
)

func main() {

	if err := config.LoadConfig(); err != nil {
		log.Fatalln(err.Error())
	}

	if len(os.Args) > 1 {
		switch (os.Args[1]) {
		case "chpasswd":
			if (config.Cfg.Ldap.Enabled) {
				fmt.Println("Not valid when LDAP is enabled")
				return
			}
			if len(os.Args) < 3 {
				fmt.Println(os.Args[0] +
					    " chpasswd <username>")
				return
			}
			fmt.Print("New Password : ")
			password, err := term.ReadPassword(0)
			fmt.Print("\n")
			if err != nil {
				log.Fatalln(err.Error())
			}
			err = db.Init(config.Cfg.Database.Type,
				      config.Cfg.Database.Url, false)
			if err != nil {
				log.Fatalln(err.Error())
			}
			defer db.Close()
			if err := db.ChangePassword(os.Args[2],
						    string(password));
			   err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(os.Args[2] + "'s password changed")
			return
		case "register":
			if (config.Cfg.Ldap.Enabled) {
				fmt.Println("Not valid when LDAP is enabled")
				return
			}
			if len(os.Args) < 3 {
				fmt.Println(os.Args[0] +
					    " register <username>")
				return
			}
			fmt.Print("Password : ")
			password, err := term.ReadPassword(0)
			fmt.Print("\n")
			if err != nil {
				log.Fatalln(err.Error())
			}
			err = db.Init(config.Cfg.Database.Type,
				      config.Cfg.Database.Url, false)
			if err != nil {
				log.Fatalln(err.Error())
			}
			defer db.Close()
			if err := db.Register(os.Args[2], string(password));
			   err != nil {
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
			err := db.Init(config.Cfg.Database.Type,
				       config.Cfg.Database.Url, false)
			if err != nil {
				log.Fatalln(err.Error())
			}
			defer db.Close()
			err = db.DeleteUser(os.Args[2])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("User " + os.Args[2] +
				    " deleted successfully")
			return
		case "init":
			err := db.Init(config.Cfg.Database.Type,
				       config.Cfg.Database.Url, true)
			if err != nil {
				log.Fatalln(err.Error())
			}
			db.Close()
			return
		case "update":
			err := db.Init(config.Cfg.Database.Type,
				       config.Cfg.Database.Url, false)
			if err != nil {
				log.Fatalln(err.Error())
			}
			db.UpdateTable()
			db.Close()
			return
		}
		fmt.Println("usage: " + os.Args[0] + " [command]")
		fmt.Println("commands :")
		fmt.Println("\tchpasswd <username> - Change user password")
		fmt.Println("\tregister <username> - Create user")
		fmt.Println("\trmuser <username> - Remove user")
		fmt.Println("\tupdate - Update database " +
			"(Warning, it is recommended to do a backup of " +
			"the database before using this command)")
		fmt.Println("\tinit - Initialize database")
		return
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := access.Init(); err != nil {
		log.Fatalln(err.Error())
	}

	if err := gmi.LoadTemplate(config.Cfg.Gemini.Templates); err != nil {
		log.Fatalln(err.Error())
	}

	err := db.Init(config.Cfg.Database.Type,
		       config.Cfg.Database.Url, false)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer db.Close()
	if err := repo.Init("repos"); err != nil {
		log.Fatalln(err.Error())
	}

	if !config.Cfg.Git.Remote.Enabled {
		go httpgit.Listen("repos/",
				  config.Cfg.Git.Address,
				  config.Cfg.Git.Port)
	}
	go auth.Decrease()

	gig.DefaultLoggerConfig.Format = "${time_rfc3339} - ${remote_ip} | " +
					 "Path=${path}, Status=${status}, " +
					 "Latency=${latency}\n"
	g := gig.Default()
	g.Use(gig.Recover())
	g.Static("/static", "./static")

	passAuth := gig.PassAuth(
		func(sig string, c gig.Context) (string, error) {
			_, b := db.GetUser(sig)
			if !b {
				return "/login", nil
			}
			if c.Param("csrf") == "" {
				csrf.New(c)
			} else if err := csrf.Verify(c); err != nil {
				return "", err
			}
			return "", nil
		})

	secure := g.Group("/account", passAuth)

	secure.Handle("", gmi.ShowAccount)
	// groups management
	secure.Handle("/groups", gmi.ShowGroups)
	secure.Handle("/groups/:group", gmi.ShowMembers)
	secure.Handle("/groups/:group/:csrf/desc", gmi.SetGroupDesc)
	secure.Handle("/groups/:group/:csrf/add", gmi.AddToGroup)
	secure.Handle("/groups/:group/:csrf/leave", gmi.LeaveGroup)
	secure.Handle("/groups/:group/:csrf/delete", gmi.DeleteGroup)
	secure.Handle("/groups/:group/:csrf/kick/:user", gmi.RmFromGroup)

	// repository settings
	secure.Handle("/repo/:repo/*", gmi.RepoFile)
	secure.Handle("/repo/:repo/:csrf/togglepublic", gmi.TogglePublic)
	secure.Handle("/repo/:repo/:csrf/chname", gmi.ChangeRepoName)
	secure.Handle("/repo/:repo/:csrf/chdesc", gmi.ChangeRepoDesc)
	secure.Handle("/repo/:repo/:csrf/delrepo", gmi.DeleteRepo)

	// access management
	secure.Handle("/repo/:repo/access", gmi.ShowAccess)
	secure.Handle("/repo/:repo/access/:csrf/add", gmi.AddUserAccess)
	secure.Handle("/repo/:repo/access/:csrf/addg", gmi.AddGroupAccess)
	secure.Handle("/repo/:repo/access/:user/:csrf/first",
			gmi.UserAccessFirstOption)
	secure.Handle("/repo/:repo/access/:user/:csrf/second",
			gmi.UserAccessSecondOption)
	secure.Handle("/repo/:repo/access/:group/g/:csrf/first",
			gmi.GroupAccessFirstOption)
	secure.Handle("/repo/:repo/access/:group/g/:csrf/second",
			gmi.GroupAccessSecondOption)
	secure.Handle("/repo/:repo/access/:user/:csrf/kick",
			gmi.RemoveUserAccess)
	secure.Handle("/repo/:repo/access/:group/g/:csrf/kick",
			gmi.RemoveGroupAccess)

	// repository view
	secure.Handle("/repo/:repo", gmi.RepoLog)
	secure.Handle("/repo/:repo/license", gmi.RepoLicense)
	secure.Handle("/repo/:repo/readme", gmi.RepoReadme)
	secure.Handle("/repo/:repo/refs", gmi.RepoRefs)
	secure.Handle("/repo/:repo/files", gmi.RepoFiles)
	secure.Handle("/repo/:repo/files/:blob", gmi.RepoFileContent)

	// user page
	secure.Handle("/:csrf/chdesc", gmi.ChangeDesc)
	secure.Handle("/:csrf/addrepo", gmi.AddRepo)
	secure.Handle("/:csrf/addgroup", gmi.AddGroup)
	// otp
	secure.Handle("/otp", gmi.ShowOTP)
	secure.Handle("/otp/:csrf/qr", gmi.CreateTOTP)
	secure.Handle("/otp/:csrf/confirm", gmi.ConfirmTOTP)
	secure.Handle("/otp/:csrf/rm", gmi.RemoveTOTP)
	// token
	secure.Handle("/token", gmi.ListTokens)
	secure.Handle("/token/:csrf/new", gmi.CreateToken)
	secure.Handle("/token/:csrf/secure", gmi.ToggleTokenAuth)
	secure.Handle("/token/:csrf/renew/:token", gmi.RenewToken)
	secure.Handle("/token/:csrf/delete/:token", gmi.DeleteToken)

	if !config.Cfg.Ldap.Enabled {
		secure.Handle("/:csrf/chpasswd", gmi.ChangePassword)
	}

	secure.Handle("/:csrf/disconnect", gmi.Disconnect)
	secure.Handle("/:csrf/disconnectall", gmi.DisconnectAll)

	if config.Cfg.Git.Key != "" {
		api := g.Group("/api")
		api.Handle("/:key/init/:username/:repo", repo.ApiInit)
		api.Handle("/:key/rm/:username/:repo", repo.ApiRemove)
		api.Handle("/:key/mv/:username/:repo/:newname",
			   repo.ApiRename)
	}

	var public *gig.Group
	if config.Cfg.Git.Public {
		public = g.Group("/repo")
	} else {
		public = g.Group("/repo", passAuth)
	}

	public.Handle("", gmi.PublicList)
	public.Handle("/:user/:repo/*", gmi.PublicFile)
	public.Handle("/:user", gmi.PublicAccount)
	public.Handle("/:user/:repo", gmi.PublicLog)
	public.Handle("/:user/:repo/refs", gmi.PublicRefs)
	public.Handle("/:user/:repo/license", gmi.PublicLicense)
	public.Handle("/:user/:repo/readme", gmi.PublicReadme)
	public.Handle("/:user/:repo/files", gmi.PublicFiles)
	public.Handle("/:user/:repo/files/:blob", gmi.PublicFileContent)

	g.PassAuthLoginHandle("/login", gmi.Login)

	if config.Cfg.Users.Registration {
		g.Handle("/register", gmi.Register)
		g.Handle("/register/:name", gmi.RegisterConfirm)
	}
	g.Handle("/otp", gmi.LoginOTP)

	g.Handle("/", func(c gig.Context) error {
		return gmi.ShowIndex(c)
	})

	err = g.Run(config.Cfg.Gemini.Address + ":" + config.Cfg.Gemini.Port,
		    config.Cfg.Gemini.Certificate, config.Cfg.Gemini.Key)

	if err != nil {
		log.Fatal(err.Error())
	}
}
