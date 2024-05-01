package csrf

import (
	"errors"
	"gemigit/db"

	"crypto/rand"
	"github.com/pitr/gig"
)

var tokens = map[string]string{}

const characters = "abcdefghijklmnopqrstuvwxyz0123456789"
func randomString(n int) string {
        var random [1024]byte
	if n > 1024 { return "" }
        b := make([]byte, n)
        rand.Read(random[:n])
        for i := range b {
                b[i] = characters[int64(random[i]) % int64(len(characters))]
        }
        return string(b)
}

func New(c gig.Context) error {
	sig := c.CertHash()
	exist := false
	if sig != "" { _, exist = db.GetUser(sig) }
	if !exist { return c.NoContent(gig.StatusRedirectTemporary, "/") }
	token := randomString(16)
	tokens[sig] = token
	return nil
}

func Verify(c gig.Context) error {
	sig := c.CertHash()
	token, exist := tokens[sig]
	if exist { _, exist = db.GetUser(sig) }
	if !exist || token != c.Param("csrf") {
		return errors.New("invalid csrf token")
	}
	return nil
}

func Token(sig string) string {
	return tokens[sig]
}
