package dandler

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/groupcache"
	"github.com/stretchr/testify/assert"
)

func init() {
	groupcache.NewHTTPPool("http://127.0.0.1:12345")
}

// func TestLoadThumbnail(t *testing.T) {
// 	testData := []struct {
// 		imageName string
// 		size      int64
// 	}{
// 		{
// 			imageName: "accidentally_save_file.gif",
// 			size:      17861,
// 		}, {
// 			imageName: "blocked_us.png",
// 			size:      44940,
// 		}, {
// 			imageName: "carlton_pls.jpg",
// 			size:      22806,
// 		}, {
// 			imageName: "lemur_pudding_cups.jpg",
// 			size:      72840,
// 		}, {
// 			imageName: "spooning_a_barret.png",
// 			size:      47306,
// 		}, {
// 			imageName: "whats_in_the_case.gif",
// 			size:      48763,
// 		},
// 	}

// 	tempdir, err := ioutil.TempDir("", "sp9k1-")
// 	if err != nil {
// 		t.Fatalf("failed creating test directory: %s", err)
// 	}

// 	h := thumbnailHandler{x: 200, y: 200, raw: "testdata", thumbExt: "png", thumbs: tempdir}

// 	for id, test := range testData {
// 		h.loadThumbnail(test.imageName)
// 		info, err := os.Stat(h.generateThumbPath(test.imageName))
// 		if err != nil {
// 			t.Logf("#%d - failed to stat thumbnail [%s] tempdir [%s]: %s",
// 				id, test.imageName, tempdir, err)
// 			t.Fail()
// 			continue
// 		}
// 		assert.Equal(t, test.size, info.Size(),
// 			"#%d [%s] - size does not match - tempDir [%s]", id, test.size, tempdir)
// 	}
// }

func TestThumbCache(t *testing.T) {
	var testData = []struct {
		uri           string
		code          int
		md5           string
		contentLength int64
		contentType   string
	}{
		{
			uri:           "/accidentally_save_file.gif",
			code:          200,
			md5:           "bc587c694204580315614011d6b702ce",
			contentLength: 25162,
			contentType:   "image/png",
		}, {
			uri:           "/blocked_us.png",
			code:          200,
			md5:           "be0261c7ed6c869e3462f1688f040ab8",
			contentLength: 66336,
			contentType:   "image/png",
		}, {
			uri:           "/carlton_pls.jpg",
			code:          200,
			md5:           "e2d15c65598dd54f0b72c118134344a3",
			contentLength: 33345,
			contentType:   "image/png",
		}, {
			uri:           "/lemur_pudding_cups.jpg",
			code:          200,
			md5:           "53070f5de5e3d2e44e6b4af461fad761",
			contentLength: 125386,
			contentType:   "image/png",
		}, {
			uri:           "/spooning_a_barret.png",
			code:          200,
			md5:           "2f53597728f846ceb39f88bf27f44d4f",
			contentLength: 71299,
			contentType:   "image/png",
		}, {
			uri:           "/whats_in_the_case.gif",
			code:          200,
			md5:           "1990381bd41ea22983e1a806d3381afa",
			contentLength: 96063,
			contentType:   "image/png",
		}, {
			uri:           "/bad.target",
			code:          500,
			md5:           "",
			contentLength: 0,
			contentType:   "",
		}, {
			uri:           "/accidentally_save_file.gif",
			code:          200,
			md5:           "bc587c694204580315614011d6b702ce",
			contentLength: 25162,
			contentType:   "image/png",
		}, {
			uri:           "/blocked_us.png",
			code:          200,
			md5:           "be0261c7ed6c869e3462f1688f040ab8",
			contentLength: 66336,
			contentType:   "image/png",
		}, {
			uri:           "/carlton_pls.jpg",
			code:          200,
			md5:           "e2d15c65598dd54f0b72c118134344a3",
			contentLength: 33345,
			contentType:   "image/png",
		},
	}

	logger := log.New(ioutil.Discard, "", 0)
	// logger := log.New(os.Stderr, "", 0)
	ts := httptest.NewServer(ThumbCache(logger, 300, 250, 64<<20, "./testdata/", "test", "png"))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {
		t.Run(fmt.Sprintf("TestThumbCache-#%d[%s]", testID, test.uri), func(t *testing.T) {
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

			assert.Equal(t, test.code, res.StatusCode, "status code does not match: ")
			if test.code != 200 {
				if res.StatusCode != test.code {
					t.Logf("the response returned: \n%#v\n", res)
				}
				return
			}
			assert.Equal(t, test.contentLength, res.ContentLength, "ContentLength does not match: ")
			assert.Equal(t, test.contentType, res.Header.Get("Content-Type"), "Content-Type does not match: ")

			body, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Error(err)
				return
			}
			assert.Equal(t, test.md5, fmt.Sprintf("%x", md5.Sum(body)), "mismatched body returned: ")
		})
	}
}

func TestThumbCache_JPG(t *testing.T) {
	var testData = []struct {
		uri           string
		code          int
		md5           string
		contentLength int64
		contentType   string
	}{
		{
			uri:           "/accidentally_save_file.gif",
			code:          200,
			md5:           "2aa9ba78ec27dc96a3f5603e9e8eb646",
			contentLength: 12489,
			contentType:   "image/",
		}, {
			uri:           "/blocked_us.png",
			code:          200,
			md5:           "2fc5189bea70182964bf9126bcb3f0be",
			contentLength: 10887,
			contentType:   "image/",
		}, {
			uri:           "/carlton_pls.jpg",
			code:          200,
			md5:           "950e11dcdbbe9e27781aed1e815ff83f",
			contentLength: 5081,
			contentType:   "image/",
		}, {
			uri:           "/lemur_pudding_cups.jpg",
			code:          200,
			md5:           "b5a688f25e0c248a6b101467957fc989",
			contentLength: 17019,
			contentType:   "image/",
		}, {
			uri:           "/spooning_a_barret.png",
			code:          200,
			md5:           "b62b31ec6cfc5fd85dec71a3592373a8",
			contentLength: 10705,
			contentType:   "image/",
		}, {
			uri:           "/whats_in_the_case.gif",
			code:          200,
			md5:           "806a2539113d46547dbc0fe779e5c4f3",
			contentLength: 7574,
			contentType:   "image/",
		}, {
			uri:           "/bad.target.png",
			code:          500,
			md5:           "",
			contentLength: 0,
			contentType:   "",
		}, {
			uri:           "/accidentally_save_file.gif",
			code:          200,
			md5:           "2aa9ba78ec27dc96a3f5603e9e8eb646",
			contentLength: 12489,
			contentType:   "image/",
		}, {
			uri:           "/blocked_us.png",
			code:          200,
			md5:           "2fc5189bea70182964bf9126bcb3f0be",
			contentLength: 10887,
			contentType:   "image/",
		}, {
			uri:           "/carlton_pls.jpg",
			code:          200,
			md5:           "950e11dcdbbe9e27781aed1e815ff83f",
			contentLength: 5081,
			contentType:   "image/",
		},
	}

	for _, ext := range []string{"jpg", "jpeg"} {
		logger := log.New(ioutil.Discard, "", 0)
		// logger := log.New(os.Stderr, "", 0)
		ts := httptest.NewServer(ThumbCache(logger, 300, 250, 64<<20, "./testdata/", "test"+ext, ext))
		defer ts.Close()

		baseURL, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("failed to parse url: %s", err)
		}

		for testID, test := range testData {
			t.Run(fmt.Sprintf("TestThumbCache[%s]-#%d-[%s]", ext, testID, test.uri), func(t *testing.T) {
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

				assert.Equal(t, test.code, res.StatusCode, "status code does not match: ")
				if test.code != 200 {
					if res.StatusCode != test.code {
						t.Logf("the response returned: \n%#v\n", res)
					}
					return
				}
				assert.Equal(t, test.contentLength, res.ContentLength, "ContentLength does not match: ")
				assert.Equal(t, test.contentType+ext, res.Header.Get("Content-Type"), "Content-Type does not match: ")

				body, err := ioutil.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					t.Error(err)
					return
				}
				assert.Equal(t, test.md5, fmt.Sprintf("%x", md5.Sum(body)), "mismatched body returned: ")
			})
		}
	}
}

func TestThumbCache_GeneratePaths(t *testing.T) {
	testData := []struct {
		imageName string
		rawPath   string
		thumbPath string
	}{
		{
			imageName: "accidentally_save_file.gif",
			rawPath:   "testdata/accidentally_save_file.gif",
		}, {
			imageName: "blocked_us.png",
			rawPath:   "testdata/blocked_us.png",
		}, {
			imageName: "carlton_pls.jpg",
			rawPath:   "testdata/carlton_pls.jpg",
		}, {
			imageName: "lemur_pudding_cups.jpg",
			rawPath:   "testdata/lemur_pudding_cups.jpg",
		}, {
			imageName: "spooning_a_barret.png",
			rawPath:   "testdata/spooning_a_barret.png",
		}, {
			imageName: "whats_in_the_case.gif",
			rawPath:   "testdata/whats_in_the_case.gif",
		},
	}

	h := thumbnailHandler{raw: "testdata", thumbs: "output", thumbExt: "jpg"}

	for id, test := range testData {
		assert.Equal(t, test.rawPath, h.generateRawPath(test.imageName), "#%d - wrong raw path", id)
	}
}
