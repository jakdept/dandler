package dandler

import (
	"bufio"
	"bytes"
	"log"
	"net/http"

	"github.com/golang/groupcache"
)

const Megabyte int = 1 << 20
const Gigabyte int = 1 << 30

const ()

type cacheHandler struct {
	l *log.Logger
	h http.Handler
	c *groupcache.Group
	o int
}

func (h thumbCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := new([]byte)
	err := h.cache.Get(nil, r.URL.Path, groupcache.AllocatingByteSliceSink(data))
	if err != nil {
		h.l.Printf("cache miss: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(*data)), r)
	if err != nil {
		h.l.Printf("error reading response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = resp.Write(w)
	if err != nil {
		h.l.Printf("error writing response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
