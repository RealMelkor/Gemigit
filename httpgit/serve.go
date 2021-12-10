package httpgit

import (
	"gemigit/db"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	githttpxfer "github.com/nulab/go-git-http-xfer/githttpxfer"
)

func Listen(path string, port int) {
	ghx, err := githttpxfer.New(path, "git")
	if err != nil {
		log.Fatalln("GitHTTPXfer instance could not be created. ", err.Error())
	}

	chain := newChain()
	chain.use(logging)
	chain.use(basicAuth)
	handler := chain.build(ghx)

	log.Println("Http server started on port", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), handler); err != nil {
		log.Fatalln("ListenAndServe: ", err.Error())
	}
}

type middleware func(http.Handler) http.Handler

func newChain() *chain {
	return &chain{[]middleware{}}
}

type chain struct {
	middlewares []middleware
}

func (c *chain) use(m middleware) {
	c.middlewares = append(c.middlewares, m)
}

func (c *chain) build(h http.Handler) http.Handler {
	for i := range c.middlewares {
		h = c.middlewares[len(c.middlewares)-1-i](h)
	}
	return h
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Println("["+r.RemoteAddr+"]["+r.Method+"]", r.URL.String(), t2.Sub(t1))
	})
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := strings.Split(r.URL.Path[1:], "/")
		if len(params) < 2 {
			renderNotFound(w)
			return
		}
		if strings.Contains(r.URL.Path, "git-upload-pack") || strings.Contains(r.URL.RawQuery, "git-upload-pack") {
			b, err := db.IsRepoPublic(params[1], params[0])
			if err != nil {
				renderNotFound(w)
				return
			}
			if b {
				next.ServeHTTP(w, r)
				return
			}
		}
		username, password, ok := r.BasicAuth()
		if !ok {
			renderUnauthorized(w)
			return
		}
		ok, err := db.CheckAuth(username, password)
		if err != nil {
			log.Println(err.Error())
		}
		if !ok || err != nil {
			renderUnauthorized(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func renderNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound)))
	w.Header().Set("Content-Type", "text/plain")
}

func renderUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Please enter your username and password."`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
	w.Header().Set("Content-Type", "text/plain")
}
