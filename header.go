package dandler

import (
	"net/http"

	"github.com/jakdept/drings"
)

// Header returns a handler that adds the given handler to the response.
func Header(name, msg string, handler http.Handler) http.Handler {
	return headerHandler{key: name, value: msg, child: handler}
}

type headerHandler struct {
	key   string
	value string
	child http.Handler
}

func (h headerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(h.key, h.value)
	h.child.ServeHTTP(w, r)
}

// ASCIIHeader adds a multiline header to the response - a header for each line.
// It is used to add ascii art to the header of a response.
//
// Both key and value can be a multiline string = and will be broken up as
// needed. Key will be repeated for the duration of value for each header line.
//
// You need to make sure both key and value have a consistent length after white space
// trimming on each line, as HTTP will do that anyway. If you don't, this will
// split on newlines, trim spaces, then right pad the key with underscores so
// things line up anyway.
func ASCIIHeader(key, value, padChar string, child http.Handler) http.Handler {
	keyArr := drings.SplitAndTrimSpace(key, "\n")
	width := drings.MaxLen(keyArr)
	keyArr = drings.PadAllRight(keyArr, padChar, width)

	valueArr := drings.SplitAndTrimSpace(key, "\n")

	return asciiHeaderHandler{key: keyArr, value: valueArr, child: child}
}

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
