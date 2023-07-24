package gmi

import (
	"gemigit/repo"
	"github.com/pitr/gig"
)

func PublicFile(c gig.Context) error {
        data, mtype, err := serveFile(c.Param("repo"), c.Param("user"),
                                      c.Param("*"))
        if err != nil {
                return c.NoContent(gig.StatusBadRequest, err.Error())
        }
        return c.Blob(mtype, data)
}

func PublicFileContent(c gig.Context) error {
	content, err := repo.GetPublicFile(c.Param("repo"), c.Param("user"),
					   c.Param("blob"))
	if err != nil {
		return c.NoContent(gig.StatusBadRequest, err.Error())
	}
	header := "=>/repo/" + c.Param("user") + "/" + c.Param("repo") +
		"/files Go Back\n\n"
	return c.Gemini(header + showFileContent(content))
}

func PublicRefs(c gig.Context) error {
	return showRepo(c, pageRefs, false)
}

func PublicLicense(c gig.Context) error {
	return showRepo(c, pageLicense, false)
}

func PublicReadme(c gig.Context) error {
	return showRepo(c, pageReadme, false)
}

func PublicLog(c gig.Context) error {
	return showRepo(c, pageLog, false)
}

func PublicFiles(c gig.Context) error {
	return showRepo(c, pageFiles, false)
}
