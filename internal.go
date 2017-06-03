//go:generate statik -src=./static

package handler

import (
	"fmt"
	"log"
	"net/http"
	"path"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/jakdept/sp9k1/statik"
)

// InternalHandler serves a static, in memory filesystem..
func InternalHandler(logger *log.Logger, fs http.FileSystem) http.Handler {
	return internalHandler{handler: http.FileServer(fs), l: logger}
}

type internalHandler struct {
	handler http.Handler
	l       *log.Logger
}

func (c internalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// block access to files ending in .template
	if path.Ext(r.URL.Path) == ".template" {
		http.Error(w, fmt.Sprintf("template requested, blocked: %s", r.URL.Path), http.StatusForbidden)
		c.l.Printf("403 - error responding: %s", r.URL.Path)
		return
	}

	c.handler.ServeHTTP(w, r)
	return
}
