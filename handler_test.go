// built with goldie
// if golden files in fixture dir are manually verified, you can update with
// go test -update

package handler

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	_ "github.com/jakdept/sp9k1/statik"
	"github.com/rakyll/statik/fs"
	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

var testFS http.FileSystem

func init() {
	var err error
	testFS, err = fs.New()
	if err != nil {
		log.Fatalf("Failed to load statik fs, aborting tests: %s", err)
	}
	goldie.FixtureDir = "testdata/fixtures"
}

func TestInternalHandler(t *testing.T) {
	// todo: re-enable this test
	log.Println("internal handler test skipped")
	t.Skip()
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
			contentType:   "text/css; charset=utf-8",
		}, {
			uri:           "/default.css",
			code:          200,
			md5:           "b1cf11f4d2cda79f08a58383863346a7",
			contentLength: 1868,
			contentType:   "text/css; charset=utf-8",
		}, {
			uri:           "/grid.js",
			code:          200,
			md5:           "c1b9a03d47a42720891989a5844e9e3c",
			contentLength: 14173,
			contentType:   "application/javascript",
		}, {
			uri:           "/modernizr.custom.js",
			code:          200,
			md5:           "3d025169b583ce5c3af13060440e2277",
			contentLength: 8281,
			contentType:   "application/javascript",
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
			code:          403,
			md5:           "23115a2a2e7d25f86bfb09392986681d",
			contentLength: 0,
			contentType:   "text/html; charset=utf-8",
		},
	}

	logger := log.New(ioutil.Discard, "", 0)
	ts := httptest.NewServer(InternalHandler(logger, testFS))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {
		t.Run(fmt.Sprintf("TestInternalHandler #%d - [%s]", testID, test.uri), func(t *testing.T) {
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
	testData := []struct {
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

func TestGeneratePaths(t *testing.T) {
	testData := []struct {
		imageName string
		rawPath   string
		thumbPath string
	}{
		{
			imageName: "accidentally_save_file.gif",
			rawPath:   "testdata/accidentally_save_file.gif",
			thumbPath: "output/accidentally_save_file.gif.jpg",
		}, {
			imageName: "blocked_us.png",
			rawPath:   "testdata/blocked_us.png",
			thumbPath: "output/blocked_us.png.jpg",
		}, {
			imageName: "carlton_pls.jpg",
			rawPath:   "testdata/carlton_pls.jpg",
			thumbPath: "output/carlton_pls.jpg.jpg",
		}, {
			imageName: "lemur_pudding_cups.jpg",
			rawPath:   "testdata/lemur_pudding_cups.jpg",
			thumbPath: "output/lemur_pudding_cups.jpg.jpg",
		}, {
			imageName: "spooning_a_barret.png",
			rawPath:   "testdata/spooning_a_barret.png",
			thumbPath: "output/spooning_a_barret.png.jpg",
		}, {
			imageName: "whats_in_the_case.gif",
			rawPath:   "testdata/whats_in_the_case.gif",
			thumbPath: "output/whats_in_the_case.gif.jpg",
		},
	}

	h := thumbnailHandler{raw: "testdata", thumbs: "output", thumbExt: "jpg"}

	for id, test := range testData {
		assert.Equal(t, test.rawPath, h.generateRawPath(test.imageName), "#%d - wrong raw path", id)
		assert.Equal(t, test.thumbPath, h.generateThumbPath(test.imageName), "#%d - wrong thumb path", id)
	}
}
