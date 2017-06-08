package dandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	ts := httptest.NewServer(Header("superheader", "secret value", Success("yay")))

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, "secret value", resp.Header.Get("superheader"))
}
