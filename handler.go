//go:generate statik -src=./static

package handler

import (
	"fmt"
	"html/template"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"image/jpeg"
	"image/png"

	"github.com/jakdept/dir"
	_ "github.com/jakdept/sp9k1/statik"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
	"github.com/traherom/memstream"
)

// SplitHandler allows the routing of one handler at /, and another at all locations below /.
func SplitHandler(root, more http.Handler) http.Handler {
	return splitHandler{bare: root, more: more}
}

type splitHandler struct {
	bare http.Handler
	more http.Handler
}

func (p splitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if path.Clean(r.URL.Path) == "/" {
		p.bare.ServeHTTP(w, r)
	} else {
		p.more.ServeHTTP(w, r)
	}
}

func DirSplitHandler(logger *log.Logger, basepath string, done <-chan struct{}, folder, other http.Handler) http.Handler {
	tracker, err := dir.Watch(basepath)
	if err != nil {
		log.Fatalf("failed to watch directory [%s] - %v", basepath, err)
	}
	go func() {
		<-done
		tracker.Close()
	}()

	return dirSplitHandler{dir: tracker, folder: folder, other: other}
}

type dirSplitHandler struct {
	dir    *dir.Tracker
	folder http.Handler
	other  http.Handler
}

func (h dirSplitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.dir.In(path.Clean(r.URL.Path)) {
		h.folder.ServeHTTP(w, r)
	} else {
		h.other.ServeHTTP(w, r)
	}
}

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

// ContentTypeHandler serves a given file back to the requester, and determines content type by algorithm only.
// It does not use the file's extension to determine the content type.
func ContentTypeHandler(logger *log.Logger, basePath string) http.Handler {
	return contentTypeHandler{basePath: basePath, l: logger}
}

type contentTypeHandler struct {
	basePath string
	l        *log.Logger
}

type errorHandler struct {
	code int
	msg  string
}

func (h errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, h.msg, h.code)
}

func ErrorHandler(code int, msg string, args ...interface{}) http.Handler {
	return errorHandler{code: code, msg: fmt.Sprintf(msg, args...)}
}
