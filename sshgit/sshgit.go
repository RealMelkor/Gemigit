package sshgit

import (
	"github.com/gliderlabs/ssh"
	"gemigit/access"
	"gemigit/config"
	"gemigit/db"
	"log"
	"strconv"
	"strings"
	"os/exec"
	"os"
	gossh "golang.org/x/crypto/ssh"
)

func handle(s ssh.Session) {

	invalidPath := []byte("Invalid path\n")
	notFound := []byte("Repository not found\n")
	forbidden := []byte("Access forbidden\n")
	readCommand := s.Command()[0] == "git-upload-pack"
	writeCommand := s.Command()[0] == "git-receive-pack"

	if (!readCommand && !writeCommand) {
		s.Stderr().Write([]byte("invalid command\n"))
		return
	}
	if len(s.Command()) < 2 {
		s.Stderr().Write(invalidPath)
		return
	}
	arg := s.Command()[1]
	if arg[0] != '/' {
		s.Stderr().Write(invalidPath)
		return
	}
	arg = arg[1:]
	if arg[len(arg) - 1] == '/' {
		arg = arg[:len(arg) - 1]
	}
	args := strings.Split(arg, "/")
	if len(args) != 2 {
		s.Stderr().Write(invalidPath)
		return
	}
	owner := args[0]
	repo := args[1]
	username := s.User()
	password := s.Context().Value("password").(string)
	readOnly := readCommand

	public := false
	if config.Cfg.Git.Public {
		var err error
		public, err = db.IsRepoPublic(repo, owner)
		if err != nil {
			s.Stderr().Write(notFound)
			log.Println(err.Error())
			return
		}
	}

	if !public || !readOnly {
		pass, err := db.CanUsePassword(repo, owner, username)
		if err != nil {
			log.Println(err.Error())
			s.Stderr().Write(forbidden)
			return
		}
		err = access.Login(username, password, true, pass, !readOnly)
		if err != nil {
			log.Println(err.Error())
			s.Stderr().Write([]byte(err.Error() + "\n"))
			return
		}
		if readOnly {
			err = access.HasReadAccess(repo, owner, username)
		} else {
			err = access.HasWriteAccess(repo, owner, username)
		}
		if err != nil {
			log.Println(err.Error())
			s.Stderr().Write(forbidden)
			return
		}
	}

	var command string
	if writeCommand {
		command = "git-receive-pack"
	} else {
		command = "git-upload-pack"
	}
	cmd := exec.Command(command,
			config.Cfg.Git.Path + "/" + owner + "/" + repo)
	cmd.Stdin = s
	cmd.Stdout = s
	if err := cmd.Run(); err != nil {
		log.Println(err)
		return
	}
}

func Listen(path string, address string, port int) {
	var server ssh.Server
	server.Handle(handle)
	server.KeyboardInteractiveHandler = func(ctx ssh.Context,
			challenge gossh.KeyboardInteractiveChallenge) bool {
		if ctx.User() == "anon" {
			ctx.SetValue("password", "")
			return true
		}
		answers, err := challenge("", "",
				[]string{"password:"}, []bool{false})
		if err != nil {
			log.Println(err)
			return false
		}
		ctx.SetValue("password", answers[0])
		return true
	}

	server.Addr = config.Cfg.Git.SSH.Address + ":" +
		strconv.Itoa(config.Cfg.Git.SSH.Port)
	server.PasswordHandler = ssh.PasswordHandler(
		func(ctx ssh.Context, password string) bool {
			ctx.SetValue("password", password)
			return true
		})
	data, err := os.ReadFile(config.Cfg.Gemini.Key)
	if err != nil {
		log.Fatalln(err)
		return
	}
	key, err := gossh.ParsePrivateKey(data)
	if err != nil {
		log.Fatalln(err)
		return
	}
	server.AddHostKey(key)
	log.Println("SSH server started on port", config.Cfg.Git.SSH.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
