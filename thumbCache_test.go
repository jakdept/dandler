package dandler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/groupcache"
	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	groupcache.NewHTTPPool("http://127.0.0.1:12345")
}

func TestThumbCache(t *testing.T) {
	var testData = []struct {
		uri         string
		code        int
		contentType string
	}{
		{
			uri:         "/accidentally_save_file.gif",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/blocked_us.png",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/carlton_pls.jpg",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/lemur_pudding_cups.jpg",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/spooning_a_barret.png",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/whats_in_the_case.gif",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/bad.target",
			code:        404,
			contentType: "",
		}, {
			uri:         "/accidentally_save_file.gif",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/blocked_us.png",
			code:        200,
			contentType: "image/png",
		}, {
			uri:         "/carlton_pls.jpg",
			code:        200,
			contentType: "image/png",
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
		t.Run(fmt.Sprintf("TestThumbCache-%d", testID), func(t *testing.T) {
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
			assert.Equal(t, test.contentType, res.Header.Get("Content-Type"), "Content-Type does not match: ")

			// body, err := ioutil.ReadAll(res.Body)
			// res.Body.Close()
			// require.NoError(t, err)
			// goldie.Assert(t, t.Name(), body)
		})
	}
}

func TestThumbCache_JPG(t *testing.T) {
	var testData = []struct {
		uri         string
		code        int
		contentType string
	}{
		{
			uri:         "/accidentally_save_file.gif",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/blocked_us.png",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/carlton_pls.jpg",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/lemur_pudding_cups.jpg",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/spooning_a_barret.png",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/whats_in_the_case.gif",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/bad.target.png",
			code:        404,
			contentType: "",
		}, {
			uri:         "/accidentally_save_file.gif",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/blocked_us.png",
			code:        200,
			contentType: "image/",
		}, {
			uri:         "/carlton_pls.jpg",
			code:        200,
			contentType: "image/",
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
			t.Run(fmt.Sprintf("TestThumbCache-%s-%d", ext, testID), func(t *testing.T) {
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
				assert.Equal(t, test.contentType+ext, res.Header.Get("Content-Type"), "Content-Type does not match: ")

				body, err := ioutil.ReadAll(res.Body)
				res.Body.Close()
				require.NoError(t, err)
				goldie.Assert(t, t.Name(), body)
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
