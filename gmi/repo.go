package gmi

import (
	"strconv"
	"strings"

        "gemigit/db"
        "gemigit/repo"
        "gemigit/csrf"
	io "gemigit/util"

        "github.com/pitr/gig"
	"github.com/gabriel-vasile/mimetype"
)

func accountRedirect(c gig.Context, after string) error {
	return c.NoContent(gig.StatusRedirectTemporary,
		"/account/" + csrf.Token(c.CertHash()) + "/" + after)
}

func showFileContent(content string) string {
	lines := strings.Split(content, "\n")
	file := ""
	for i, line := range lines {
		file += strconv.Itoa(i) + "\t" + line + "\n"
	}
	return strings.Replace(file, "%", "%%", -1)
}

func serveFile(name string, user string, file string) ([]byte, string, error) {
	repofile, err := repo.GetFile(name, user, file)
	if err != nil {
		return nil, "", err
	}
	reader, err := repofile.Reader()
	if err != nil {
		return nil, "", err
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}
	mtype := mimetype.Detect(buf)
	return buf, mtype.String(), nil
}

// Private

func RepoFiles(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	query, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if query == "" {
		return showRepo(c, pageFiles, true)
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
}

func RepoFileContent(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	content, err := repo.GetPrivateFile(c.Param("repo"), user.Name,
					    c.Param("blob"), c.CertHash())
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	header := "=>/account/repo/" + c.Param("repo") + "/files Go Back\n\n"
	return c.Gemini(header + showFileContent(content))
}

func RepoFile(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	data, mtype, err := serveFile(c.Param("repo"), user.Name, c.Param("*"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return c.Blob(mtype, data)
}

func TogglePublic(c gig.Context) error {
	user, exist := db.GetUser(c.CertHash())
	if !exist {
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	if err := user.TogglePublic(c.Param("repo"), c.CertHash());
	   err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accountRedirect(c, "repo/" + c.Param("repo"))
}

func ChangeRepoName(c gig.Context) error {
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
	// should check if repo exist and if the new name is free
	if err := repo.ChangeRepoDir(c.Param("repo"), user.Name, newname);
	   err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if err := user.ChangeRepoName(c.Param("repo"), newname, c.CertHash());

	   err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accountRedirect(c, "repo/" + newname)
}

func ChangeRepoDesc(c gig.Context) error {
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
		return c.NoContent(gig.StatusBadRequest, "Invalid username")
	}
	if err := user.ChangeRepoDesc(c.Param("repo"), newdesc);
	   err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accountRedirect(c, "repo/" + c.Param("repo"))
}

func DeleteRepo(c gig.Context) error {
	name, err := c.QueryString()
	if err != nil {
		return c.NoContent(gig.StatusBadRequest,
				   "Invalid input received")
	}
	if name == "" {
		return c.NoContent(gig.StatusInput,
				"To confirm type the repository name")
	}
	if name != c.Param("repo") {
		return accountRedirect(c, "repo/" + c.Param("repo"))
	}
	user, b := db.GetUser(c.CertHash())
	if !b {
		return c.NoContent(gig.StatusBadRequest,
				   "Cannot find username")
	}
	// check if repo exist
	if err := repo.RemoveRepo(name, user.Name);
	   err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	if err := user.DeleteRepo(name, c.CertHash());
	   err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	return accountRedirect(c, "")
}

func RepoRefs(c gig.Context) error {
	return showRepo(c, pageRefs, true)
}

func RepoLicense(c gig.Context) error {
	return showRepo(c, pageLicense, true)
}

func RepoReadme(c gig.Context) error {
	return showRepo(c, pageReadme, true)
}

func RepoLog(c gig.Context) error {
	return showRepo(c, pageLog, true)
}
