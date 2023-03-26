package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"gemigit/access"
	"gemigit/auth"
	"gemigit/config"
	"gemigit/db"
	"gemigit/httpgit"
	"gemigit/repo"
	"gemigit/gmi"

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
			password, err := terminal.ReadPassword(0)
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
			password, err := terminal.ReadPassword(0)
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
			defer db.Close()
			return
		}
		fmt.Println("usage: " + os.Args[0] + " [command]")
		fmt.Println("commands :")
		fmt.Println("\tchpasswd <username> - Change user password")
		fmt.Println("\tregister <username> - Create user")
		fmt.Println("\trmuser <username> - Remove user")
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
			return "", nil
		})

	secure := g.Group("/account", passAuth)

	secure.Handle("", gmi.ShowAccount)
	// groups management
	secure.Handle("/groups", gmi.ShowGroups)
	secure.Handle("/groups/:group", gmi.ShowMembers)
	secure.Handle("/groups/:group/desc", gmi.SetGroupDesc)
	secure.Handle("/groups/:group/add", gmi.AddToGroup)
	secure.Handle("/groups/:group/leave", gmi.LeaveGroup)
	secure.Handle("/groups/:group/delete", gmi.DeleteGroup)
	secure.Handle("/groups/:group/kick/:user", gmi.RmFromGroup)

	// repository settings
	secure.Handle("/repo/:repo/*", gmi.RepoFile)
	secure.Handle("/repo/:repo/togglepublic", gmi.TogglePublic)
	secure.Handle("/repo/:repo/chname", gmi.ChangeRepoName)
	secure.Handle("/repo/:repo/chdesc", gmi.ChangeRepoDesc)
	secure.Handle("/repo/:repo/delrepo", gmi.DeleteRepo)

	// access management
	secure.Handle("/repo/:repo/access", gmi.ShowAccess)
	secure.Handle("/repo/:repo/access/add", gmi.AddUserAccess)
	secure.Handle("/repo/:repo/access/addg", gmi.AddGroupAccess)
	secure.Handle("/repo/:repo/access/:user/first",
		      gmi.UserAccessFirstOption)
	secure.Handle("/repo/:repo/access/:user/second",
		      gmi.UserAccessSecondOption)
	secure.Handle("/repo/:repo/access/:group/g/first",
		      gmi.GroupAccessFirstOption)
	secure.Handle("/repo/:repo/access/:group/g/second",
		      gmi.GroupAccessSecondOption)
	secure.Handle("/repo/:repo/access/:user/kick",
		      gmi.RemoveUserAccess)
	secure.Handle("/repo/:repo/access/:group/g/kick",
		      gmi.RemoveGroupAccess)

	// repository view
	secure.Handle("/repo/:repo", gmi.RepoLog)
	secure.Handle("/repo/:repo/license", gmi.RepoLicense)
	secure.Handle("/repo/:repo/readme", gmi.RepoReadme)
	secure.Handle("/repo/:repo/refs", gmi.RepoRefs)
	secure.Handle("/repo/:repo/files", gmi.RepoFiles)
	secure.Handle("/repo/:repo/files/:blob", gmi.RepoFileContent)

	// user page
	secure.Handle("/chdesc", gmi.ChangeDesc)
	secure.Handle("/addrepo", gmi.AddRepo)
	secure.Handle("/addgroup", gmi.AddGroup)

	if !config.Cfg.Ldap.Enabled {
		secure.Handle("/chpasswd", gmi.ChangePassword)
	}

	secure.Handle("/disconnect", gmi.Disconnect)
	secure.Handle("/disconnectall", gmi.DisconnectAll)

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

	g.Handle("/", func(c gig.Context) error {
		return gmi.ShowIndex(c)
	})

	err = g.Run(config.Cfg.Gemini.Address + ":" + config.Cfg.Gemini.Port,
		    config.Cfg.Gemini.Certificate, config.Cfg.Gemini.Key)

	if err != nil {
		log.Fatal(err.Error())
	}
}
