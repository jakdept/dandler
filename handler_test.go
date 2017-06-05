package dandler

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/jakdept/sp9k1/statik"
	"github.com/stretchr/testify/assert"
)

func TestContentTypeHandler(t *testing.T) {
	var testData = []struct {
		uri           string
		code          int
		md5           string
		contentLength int64
		contentType   string
	}{
		{
			uri:           "/component.css",
			code:          200,
			md5:           "6929ee1f5b86c6e5669334b34e8fea65",
			contentLength: 3548,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/default.css",
			code:          200,
			md5:           "b1cf11f4d2cda79f08a58383863346a7",
			contentLength: 1868,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/grid.js",
			code:          200,
			md5:           "c1b9a03d47a42720891989a5844e9e3c",
			contentLength: 14173,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/modernizr.custom.js",
			code:          200,
			md5:           "3d025169b583ce5c3af13060440e2277",
			contentLength: 8281,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/page.html",
			code:          200,
			md5:           "9676bd8257ddcd3aa6a4e50a6068a3f8",
			contentLength: 5607,
			contentType:   "text/html; charset=utf-8",
		}, {
			uri:           "/bad.target",
			code:          404,
			md5:           "",
			contentLength: 0,
			contentType:   "",
		}, {
			uri:           "/page.template",
			code:          200,
			md5:           "23115a2a2e7d25f86bfb09392986681d",
			contentLength: 1503,
			contentType:   "text/html; charset=utf-8",
		}, {
			uri:           "/lemur_pudding_cups.jpg",
			code:          200,
			md5:           "f805ae46588af757263407301965c6a0",
			contentLength: 41575,
			contentType:   "image/jpeg",
		}, {
			uri:           "/spooning_a_barret.png",
			code:          200,
			md5:           "09d8be7d937b682447348acdc38c5895",
			contentLength: 395134,
			contentType:   "image/png",
		}, {
			uri:           "/whats_in_the_case.gif",
			code:          200,
			md5:           "2f43a5317fa2f60dbf32276faf3f139a",
			contentLength: 32933853,
			contentType:   "image/gif",
			// it's vitally important that we can serve files with the WRONG extension correctly
		}, {
			uri:           "/accidentally_save_file.gif",
			code:          200,
			md5:           "a305f39d197dce79acae597e81e22bf4",
			contentLength: 187967,
			contentType:   "image/png",
		}, {
			uri:           "/blocked_us.png",
			code:          200,
			md5:           "bc2272b02e6fab9c0c48d4743d4aae7e",
			contentLength: 45680,
			contentType:   "image/jpeg",
		}, {
			uri:           "/carlton_pls.jpg",
			code:          200,
			md5:           "7c0dc59a6ebad1645fca205f701edb39",
			contentLength: 871029,
			contentType:   "image/gif",
		},
	}

	logger := log.New(ioutil.Discard, "", 0)
	ts := httptest.NewServer(ContentTypeHandler(logger, "./testdata/"))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {
		t.Run(fmt.Sprintf("TestContentTypeHandle #%d - [%s]", testID, test.uri), func(t *testing.T) {
			// t.Parallel()
			uri, err := url.Parse(test.uri)
			if err != nil {
				t.Errorf("bad URI path: [%s]", test.uri)
				return
			}

			res, err := http.Get(baseURL.ResolveReference(uri).String())
			if err != nil {
				t.Error(err)
				return
			}

			assert.Equal(t, test.code, res.StatusCode, "status code does not match")
			if test.code != 200 {
				if res.StatusCode != test.code {
					t.Logf("the response returned: \n%#v\n", res)
				}
				return
			}
			assert.Equal(t, test.contentLength, res.ContentLength, "ContentLength does not match")
			assert.Equal(t, test.contentType, res.Header.Get("Content-Type"), "Content-Type does not match")

			body, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Error(err)
				return
			}
			assert.Equal(t, test.md5, fmt.Sprintf("%x", md5.Sum(body)), "mismatched body returned")
		})
	}
}

// handler will always response with 200 ok and the given body
func foundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ohai mr client")
	})
}

func TestSplitHandler(t *testing.T) {
	testData := []struct {
		uri  string
		code int
	}{
		{uri: "/", code: 200},
		{uri: "/other", code: 404},
		{uri: "", code: 200},
		{uri: "./", code: 200},
		{uri: "bad/url", code: 404},
	}

	// setup a handler that returns one thing on the main path, and another on other paths
	ts := httptest.NewServer(SplitHandler(foundHandler(), http.NotFoundHandler()))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {
		t.Run(fmt.Sprintf("TestContentTypeHandle #%d - [%s]", testID, test.uri), func(t *testing.T) {
			uri, err := url.Parse(test.uri)
			if err != nil {
				t.Errorf("bad URI path: [%s]", test.uri)
				return
			}

			res, err := http.Get(baseURL.ResolveReference(uri).String())
			if err != nil {
				t.Error(err)
				return
			}
			if res.StatusCode == 200 {
				res.Body.Close()
			}

			assert.Equal(t, test.code, res.StatusCode, "#%d - not routed properly", testID)
		})
	}
}

func TestDirSplitHandler(t *testing.T) {
	testdata := []struct {
		uri  string
		code int
	}{
		{uri: "/", code: 200},
		{uri: "/edat", code: 200},
		{uri: "/jim", code: 200},
		{uri: "/taes", code: 200},
		{uri: "", code: 200},
		{uri: "./", code: 200},
		{uri: "/other", code: 404},
		{uri: "bad/url", code: 404},
	}

	// setup a handler that returns one thing on the main path, and another on other paths
	done := make(chan struct{})
	defer close(done)
	logger := log.New(ioutil.Discard, "", 0)
	ts := httptest.NewServer(DirSplitHandler(logger, "testdata/sample_images", done,
		foundHandler(), http.NotFoundHandler()))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testdata {
		t.Run(fmt.Sprintf("TestContentTypeHandle #%d - [%s]", testID, test.uri), func(t *testing.T) {
			uri, err := url.Parse(test.uri)
			if err != nil {
				t.Errorf("bad URI path: [%s]", test.uri)
				return
			}

			res, err := http.Get(baseURL.ResolveReference(uri).String())
			if err != nil {
				t.Error(err)
				return
			}
			if res.StatusCode == 200 {
				res.Body.Close()
			}

			assert.Equal(t, test.code, res.StatusCode, "#%d - not routed properly", testID)
		})
	}
}

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

	child := ResponseCodeHandler(200, "successful %s", "response")

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

			fmt.Printf("test #%d [%s]\n%#v\n\n", id, ts.URL, resp)

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
