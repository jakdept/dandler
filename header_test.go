package dandler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	ts := httptest.NewServer(ASCIIHeader("gopher\ngolang\ngo", GolangGopherASCII, "+", Success("hello gopher")))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if assert.NoError(t, err) {
		t.FailNow()
	}
	defer resp.Body.Close()

	output, err := ioutil.ReadAll(resp.Body)
	if assert.NoError(t, err) {
		t.FailNow()
	}

	fmt.Printf("%s", output)
}
