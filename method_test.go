package dandler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodHandler_Nil(t *testing.T) {
	ts := httptest.NewServer(MethodHandler{})
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "unconfigured method\n", string(b))
}

func TestMethodHandler(t *testing.T) {
	srv := httptest.NewServer(MethodHandler{
		"GET": Success("ohai mr client"),
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "ohai mr client", string(b))

	otherResp, err := http.Post(srv.URL, "text/plain", nil)
	assert.NoError(t, err)
	defer otherResp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, otherResp.StatusCode)

	b, err = ioutil.ReadAll(otherResp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "unconfigured method\n", string(b))
}
