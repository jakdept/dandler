package dandler

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/jakdept/dir"
)

// Split allows the routing of one handler at /, and another at all locations below /.
func Split(root, more http.Handler) http.Handler {
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

// DirSplit takes two child handlers - one for the directories, one for
// the other locations. It then watches a directory - and sub directories - and
// any request that would match a directory relative to that path is routed to
// the handler for directories. All other requests are routed to the other handler.
func DirSplit(logger *log.Logger, basepath string, done <-chan struct{}, folder, other http.Handler) http.Handler {
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

// ContentType serves a given file back to the requester, and determines content type by algorithm only.
// It does not use the file's extension to determine the content type.
func ContentType(logger *log.Logger, basePath string) http.Handler {
	return contentTypeHandler{basePath: basePath, l: logger}
}

type contentTypeHandler struct {
	basePath string
	l        *log.Logger
}

// contentTypeHandler.ServeHTTP satasifies the Handler interface.
func (c contentTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(c.basePath, r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
		c.l.Printf("404 - could not open file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - could not stat file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	chunk := make([]byte, 512)

	_, err = f.Read(chunk)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - could not read from file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - could not seek within file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(chunk))
	http.ServeContent(w, r, r.URL.Path, stat.ModTime(), f)

	return
}

type responseCodeHandler struct {
	code int
	msg  string
}

func (h responseCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, h.msg, h.code)
}

// ResponseCode responds with an error to any request with the error code and message provided.
func ResponseCode(code int, msg string, args ...interface{}) http.Handler {
	return responseCodeHandler{code: code, msg: fmt.Sprintf(msg, args...)}
}

type successHandler struct {
	msg string
}

func (h successHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.msg))
}

// Success returns a handler that responds to every request with a 200 and
// always the same message.
func Success(msg string) http.Handler {
	return successHandler{msg: msg}
}

type expiresHandler struct {
	minCache  time.Duration
	maxLength time.Duration
	child     http.Handler
}

func (h expiresHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	duration := int(h.minCache.Seconds())
	if h.maxLength > 0 {
		duration += rand.Intn(int(h.maxLength.Seconds()))
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", duration))

	h.child.ServeHTTP(w, r)
}

func Expires(maxAge time.Duration, child http.Handler) http.Handler {
	return expiresHandler{minCache: maxAge, maxLength: 0, child: child}
}

func ExpiresRange(min, max time.Duration, child http.Handler) http.Handler {
	return expiresHandler{minCache: min, maxLength: max - min, child: child}
}
