package dandler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanocialHostHandler(t *testing.T) {
	testdata := []struct {
		options        int
		host           string
		port           string
		expectedStatus int
		expectedHost   string
		expectedPort   string
		expectedScheme string
	}{
		{ // test 0
			options:        0,
			host:           "",
			port:           "",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 1
			options:        ForceHost,
			host:           "desthost.com",
			port:           "",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "desthost.com",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 2
			options:        ForcePort,
			host:           "",
			port:           "1234",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "",
			expectedPort:   "1234",
			expectedScheme: "",
		}, { // test 3
			options:        ForceHTTP,
			host:           "desthost.com",
			port:           "",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 4
			options:        ForceHTTPS,
			host:           "desthost.com",
			port:           "",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "https",
		}, { // test 5
			options:        ForceHost | ForcePort,
			host:           "desthost.com",
			port:           "1234",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "desthost.com",
			expectedPort:   "1234",
			expectedScheme: "",
		}, { // test 6
			options:        ForceHTTPS | ForceTemporary,
			host:           "",
			port:           "",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "https",
		}, { // test 7
			options:        ForceHost,
			host:           "127.0.0.1",
			port:           "",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 8
			options:        ForceHTTP,
			host:           "",
			port:           "",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "http",
		},
	}

	child := SuccessHandler("child")

	for id, test := range testdata {
		t.Run(fmt.Sprintf("canocial test %d", id), func(t *testing.T) {
			// t.Parallel()

			c := http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			// create the server for the parallel test
			ts := httptest.NewServer(CanocialHostHandler(test.host, test.port, test.options, child))
			resp, err := c.Get(ts.URL)
			assert.Nil(t, err)

			// verify the response code
			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			// if it's a redirect, check the stuff
			var respURL *url.URL
			if resp.StatusCode == 307 || resp.StatusCode == 308 {
				respURL, err = url.Parse(resp.Header.Get("Location"))
				assert.Nil(t, err)

				if test.expectedScheme != "" {
					assert.Equal(t, test.expectedScheme, respURL.Scheme)
				}

				if test.expectedHost != "" || test.expectedPort != "" {
					// split the host up
					var host, port string
					if strings.Contains(respURL.Host, ":") {
						host = strings.Split(respURL.Host, ":")[0]
						port = strings.Split(respURL.Host, ":")[1]
					} else {
						host = respURL.Host
						if respURL.Scheme == "http" {
							port = "80"
						} else {
							port = "443"
						}
					}

					if test.expectedHost != "" {
						assert.Equal(t, test.expectedHost, host)
					}
					if test.expectedPort != "" {
						assert.Equal(t, test.expectedPort, port)
					}
				}
			}
		})
	}
}
