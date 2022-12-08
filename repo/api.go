package repo

import (
	"errors"
	"log"
	"gemigit/config"

	"github.com/pitr/gig"

	"bufio"
	"crypto/tls"
	"strconv"
	"strings"
)

func request(url string) error {
	conn, err := tls.Dial("tcp", config.Cfg.Git.Remote.Address + ":1965",
			      &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.Write([]byte("gemini://" + config.Cfg.Git.Remote.Address + "/" +
			  url + "\r\n"))

	reader := bufio.NewReader(conn)
	responseHeader, err := reader.ReadString('\n')
	parts := strings.Fields(responseHeader)
	status, err := strconv.Atoi(parts[0][0:1])
	meta := parts[1]
	if status == 20 {
		return nil
	}
	log.Println(parts)
	return errors.New(meta)
}

func ApiInit(c gig.Context) error {
	if c.Param("key") != config.Cfg.Git.Remote.Key {
		return c.NoContent(gig.StatusBadRequest, "Invalid key")
	}
	err := InitRepo(c.Param("repo"), c.Param("username"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
        return c.Gemini("success")
}

func ApiRemove(c gig.Context) error {
	if c.Param("key") != config.Cfg.Git.Remote.Key {
		return c.NoContent(gig.StatusBadRequest, "Invalid key")
	}
	err := RemoveRepo(c.Param("repo"), c.Param("username"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
        return c.Gemini("success")
}

func ApiRename(c gig.Context) error {
	if c.Param("key") != config.Cfg.Git.Remote.Key {
		return c.NoContent(gig.StatusBadRequest, "Invalid key")
	}
	err := ChangeRepoDir(c.Param("repo"), c.Param("username"),
			     c.Param("newname"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
        return c.Gemini("success")
}
