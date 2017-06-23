package dandler

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalHostHandler(t *testing.T) {
	testdata := []struct {
		options        int
		url            string
		expectedStatus int
		expectedHost   string
		expectedPort   string
		expectedScheme string
	}{
		{ // test 0
			options:        0,
			url:            "",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 1
			options:        ForceHost,
			url:            "desthost.com",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "desthost.com",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 2
			options:        ForcePort,
			url:            "127.0.0.1:1234",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "",
			expectedPort:   "1234",
			expectedScheme: "",
		}, { // test 3
			options:        ForceHTTP,
			url:            "desthost.com",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 4
			options:        ForceHTTPS,
			url:            "desthost.com",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "https",
		}, { // test 5
			options:        ForceHost | ForcePort,
			url:            "desthost.com:1234",
			expectedStatus: http.StatusPermanentRedirect,
			expectedHost:   "desthost.com",
			expectedPort:   "1234",
			expectedScheme: "",
		}, { // test 6
			options:        ForceHTTPS | ForceTemporary,
			url:            "",
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "https",
		}, { // test 7
			options:        ForceHost,
			url:            "127.0.0.1",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "",
		}, { // test 8
			options:        ForceHTTP,
			url:            "",
			expectedStatus: 200,
			expectedHost:   "",
			expectedPort:   "",
			expectedScheme: "http",
		},
	}

	child := Success("child")

	for id, test := range testdata {
		t.Run(fmt.Sprintf("canonical test %d", id), func(t *testing.T) {
			// t.Parallel()

			c := http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			// create the server for the parallel test
			ts := httptest.NewServer(CanonicalHostHandler(test.url, test.options, child))
			resp, err := c.Get(ts.URL)
			if !assert.NoError(t, err) {
				log.Println(err)
				t.FailNow()
			}

			// verify teh response code
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
