package dandler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jakdept/dir"
)

// IndexHandler lists all files in a directory, and passes them to template execution to build a directory listing.
func IndexHandler(logger *log.Logger, basepath string, done <-chan struct{}, templ *template.Template) http.Handler {
	tracker, err := dir.Watch(basepath)
	if err != nil {
		logger.Printf("failed to watch directory [%s] - %v", basepath, err)
		return ResponseCodeHandler(500, "failed to initialize IndexHandler - %v", err)
	}
	go func() {
		<-done
		tracker.Close()
	}()

	return indexHandler{basePath: basepath, templ: templ, l: logger, dir: tracker, done: done}
}

// This is the struct passed to the template used with an IndexHandler
type IndexData struct {
	Files []string
	Dirs  []string
}

type indexHandler struct {
	l        *log.Logger
	done     <-chan struct{}
	dir      *dir.Tracker
	basePath string
	templ    *template.Template
}

func (c indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(c.basePath, r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
		c.l.Printf("404 - could not find file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read target: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - could not stat file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	if !stat.IsDir() {
		http.Error(w, fmt.Sprintf("cannot read target: %s", r.URL.Path), http.StatusForbidden)
		c.l.Printf("403 - could not stat file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	contents, err := f.Readdir(0)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read directory: %s", r.URL.Path), http.StatusForbidden)
		c.l.Printf("403 - could not read file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	var data IndexData
	data.Dirs = c.dir.List()

	for _, each := range contents {
		if !each.IsDir() {
			// suppress directories
			if !strings.HasPrefix(each.Name(), ".") {
				// suppress hidden files
				data.Files = append(data.Files, path.Join(r.URL.Path, each.Name()))
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = c.templ.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("error building response: %s", r.URL.Path), http.StatusInternalServerError)
		c.l.Printf("500 - error responding: %s", err)
		return
	}
}
