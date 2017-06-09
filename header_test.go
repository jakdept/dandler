package dandler

import (
	"fmt"
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
	// this test may fail if the sun shines the wrong way. Really it's a crapshoot.
	if testing.Short() {
		t.Skip("skipping a finicky test")
	}
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
	for i := 0; i < len(resp.Header["Gopher"]) ||
		i < len(resp.Header["Golang"]) || i < len(resp.Header["Go++++"]); i++ {
		if i < len(resp.Header["Gopher"]) {
			headerMap = append(headerMap, resp.Header["Gopher"][i])
		}
		if i < len(resp.Header["Golang"]) {
			headerMap = append(headerMap, resp.Header["Golang"][i])
		}
		if i < len(resp.Header["Go++++"]) {
			headerMap = append(headerMap, resp.Header["Go++++"][i])
		}
	}

	rebuilt := strings.Join(headerMap, "\n")

	fmt.Println(rebuilt)
	assert.Equal(t, strings.TrimSpace(GolangGopherASCII), rebuilt)
}
