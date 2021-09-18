package dandler

import (
	"fmt"
	"log"
	"net/http"
	"path"
)

// Internal serves a static, in memory filesystem. The filesystem to be
// served should have been generated with embed or similar..
//
// Note - this is likely a duplicate of http.FileServer. It will likely be removed.
func Internal(logger *log.Logger, fs http.FileSystem) http.Handler {
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
