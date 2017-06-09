package dandler

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	ts := httptest.NewServer(Header("superheader", "secret value", Success("yay")))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, "secret value", resp.Header.Get("superheader"))
}

func TestASCIIHeader(t *testing.T) {
	handler := Success("hello gopher")
	handler = ASCIIHeader("gopher\ngolang\ngo", GolangGopherASCII, "+", handler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer resp.Body.Close()

	var headerMap []string
	for key, value := range resp.Header {
		headerMap = append(headerMap, fmt.Sprintf("%s:%s", key, value))
	}

	log.Println(strings.Join(headerMap, "++\n"))
}
