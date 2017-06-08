//go:generate statik -src=./testdata

package dandler

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
)

// Internal serves a static, in memory filesystem. The filesystem to be
// served should have been generated with github.com/rakyll/statik. It should
// also be imported directly into the current package.
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

// AsciiHeader adds a multiline header to the response - a header for each line.
// It is used to add ascii art to the header of a response.
//
// Both key and value can be a multiline string = and will be broken up as
// needed. Key will be repeated for the duration of value for each header line.
//
// You need to make sure both key and value have a consistent length after white space
// trimming on each line, as HTTP will do that anyway. If you don't, this will
// split on newlines, trim spaces, then right pad the key with underscores so
// things line up anyway.
func AsciiHeader(key, value string, child http.Handler) http.Handler {
	// eventually move this stuff out into jakdept/drings
	valueArr := strings.Split(value, "\n")
	for i := 0; i < len(valueArr); {
		valueArr[i] = strings.TrimSpace(valueArr[i])
		if valueArr[i] != "" {
			// either it's not empty and increment
			i++
		} else {
			// or eliminate this cell because it's empty
			if i < len(valueArr)-1 {
				valueArr = append(valueArr[:i], valueArr[i+1])
			} else {
				valueArr = valueArr[:i]
			}
		}
	}

	keyArr := strings.Split(key, "\n")
	maxKey := 0
	for i := 0; i < len(keyArr); {
		keyArr[i] = strings.TrimSpace(keyArr[i])
		if len(keyArr[i]) > maxKey {
			maxKey = len(keyArr[i])
		}
		if keyArr[i] != "" {
			// either it's not empty and increment
			i++
		} else {
			// or eliminate this cell because it's empty
			if i < len(keyArr)-1 {
				keyArr = append(keyArr[:i], keyArr[i+1])
			} else {
				keyArr = keyArr[:i]
			}
		}
	}

}

// AsciiPadCharacter is the character used to pad the keys in AsciiHeader
var AsciiPadCharacter = " "

type asciiHeaderHandler struct {
	key   []string
	value []string
	child http.Handler
}

func (h asciiHeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for keyCount, valueCount := 0, 0; valueCount < len(h.value); {
		keyCount++
		valueCount++
		if keyCount >= len(h.key) {
			keyCount = 0
		}
		w.Header().Add(h.key[keyCount], h.value[valueCount])
	}
	h.child.ServeHTTP(w, r)
}
