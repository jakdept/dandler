package dandler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectURIHandler(t *testing.T) {
	tests := []struct {
		tls        bool
		domain     string
		prefixURI  string
		code       int
		requestURI string
		target     string
	}{
		{
			tls:        false,
			domain:     "b.com",
			prefixURI:  "/prefix/",
			code:       301,
			requestURI: "/test",
			target:     "http://b.com/prefix/test",
		}, {
			tls:        false,
			domain:     "anotherdomain.com",
			prefixURI:  "",
			code:       302,
			requestURI: "/test",
			target:     "http://anotherdomain.com/test",
		}, {
			tls:        true,
			domain:     "secure.domain",
			prefixURI:  "",
			code:       302,
			requestURI: "/idksomething",
			target:     "https://secure.domain/idksomething",
		},
	}
	for id, tt := range tests {
		t.Run(fmt.Sprintf("test #%d", id), func(t *testing.T) {
			rh := &RedirectURIHandler{
				TLS:       tt.tls,
				Domain:    tt.domain,
				PrefixURI: tt.prefixURI,
				Code:      tt.code,
			}
			srv := httptest.NewServer(rh)
			c := srv.Client()
			c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			resp, err := c.Get(srv.URL + tt.requestURI)
			_ = resp
			_ = err
			require.NoError(t, err)
			assert.Equal(t, tt.code, resp.StatusCode)
			if tt.target != "" {
				assert.Equal(t, tt.target, resp.Header.Get("Location"))
			}
		})
	}
}
