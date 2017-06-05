package dandler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/jakdept/dir"
	_ "github.com/jakdept/sp9k1/statik"
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

// ContentTypeHandler serves a given file back to the requester, and determines content type by algorithm only.
// It does not use the file's extension to determine the content type.
func ContentTypeHandler(logger *log.Logger, basePath string) http.Handler {
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

func ResponseCodeHandler(code int, msg string, args ...interface{}) http.Handler {
	return responseCodeHandler{code: code, msg: fmt.Sprintf(msg, args...)}
}

type canocialHostHandler struct {
	host    string
	port    string
	options int
	child   http.Handler
}

const (
	ForceHTTP      = 1 << iota // force http as the redirect target
	ForceHTTPS                 // force https as the redirect target
	ForceHost                  // force the given hostname as the redirect target
	ForcePort                  // force a given port for the redirect target
	ForceTemporary             // Use a 302 for the redirect
)

func CanocialHostHandler(host, port string, options int, childHandler http.Handler) http.Handler {
	return canocialHostHandler{host: host, port: port, options: options, child: childHandler}
}

func (h canocialHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.checkHostAndPort(*r.URL) || h.checkScheme(*r.URL) {
		if h.options&ForceTemporary != 0 {
			http.Redirect(w, r, h.buildRedirect(*r.URL), http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, h.buildRedirect(*r.URL), http.StatusPermanentRedirect)
		}
	}
	h.child.ServeHTTP(w, r)
}

func (h canocialHostHandler) checkHostAndPort(url url.URL) bool {
	if strings.Contains(url.Host, ":") {
		chunks := strings.SplitN(url.Host, ":", 2)
		if h.options&ForceHost != 0 && chunks[0] != h.host {
			return true
		}
		if h.options&ForcePort != 0 && chunks[1] != h.port {
			return true
		}
	} else {
		if h.options&ForceHost != 0 && url.Host != h.host {
			return true
		}
	}
	return false
}

func (h canocialHostHandler) checkScheme(url url.URL) bool {
	switch {
	case h.options&ForceHTTPS != 0:
		return url.Scheme != "https"
	case h.options&ForceHTTP != 0:
		return url.Scheme != "http"
	default:
		return false
	}
}

func (h canocialHostHandler) buildRedirect(url url.URL) string {
	// if host or port is forced, I have to modify the host header
	if h.options&(ForceHost|ForcePort) != 0 {
		var host, port string
		if strings.Contains(url.Host, ":") {
			chunks := strings.SplitN(url.Host, ":", 2)
			host, port = chunks[0], chunks[1]
		} else {
			host = url.Host
		}
		if h.options&ForceHost != 0 {
			host = h.host
		}
		if h.options&ForcePort != 0 {
			port = h.port
		}
		if port == "" {
			url.Host = host
		} else {
			url.Host = host + ":" + port
		}
	}

	// if forcing http, change it now
	if h.options&ForceHTTP != 0 {
		url.Scheme = "http"
	}

	// if forcing https, change it now
	if h.options&ForceHTTPS != 0 {
		url.Scheme = "https"
	}

	return url.String()
}
