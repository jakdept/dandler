// built with goldie
// if golden files in fixture dir are manually verified, you can update with
// go test -update

package dandler

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jakdept/sp9k1/statik"
	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

func init() {
	goldie.FixtureDir = "testdata/fixtures"
}

func TestIndexHandler_successful(t *testing.T) {
	templateString := `{
	"files":{
		{{ range $index, $value := .Files -}}
		{{- if $index }}, 
		{{ end -}}
		"{{- . -}}"
		{{- end }}
	},
	"dirs":{
		{{ range $index, $value := .Dirs -}}
		{{- if $index }}, 
		{{ end -}}
		"{{- . -}}"
		{{- end }}
	}
}`

	testTempl := template.Must(template.New("test").Parse(templateString))

	done := make(chan struct{})
	defer close(done)

	logger := log.New(ioutil.Discard, "", 0)
	ts := httptest.NewServer(IndexHandler(logger, "testdata/sample_images", done, testTempl))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response: [%s]", err)
	}
	res.Body.Close()

	goldie.Assert(t, "templateHandler", actual)
}

// test to make sure a bad folder kicks a 404
func TestIndexHandler_badpath(t *testing.T) {
	templateString := ""
	testTempl := template.Must(template.New("test").Parse(templateString))

	done := make(chan struct{})
	defer close(done)

	logger := log.New(ioutil.Discard, "", 0)
	ts := httptest.NewServer(IndexHandler(logger, "not-a-folder", done, testTempl))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 500, res.StatusCode, "got wrong response")
}

// test to make sure that a bad template kicks a 500
func TestIndexHandler_badtemplate(t *testing.T) {
	templateString := "{{ .ValueNotPresent }}"
	testTempl := template.Must(template.New("test").Parse(templateString))

	done := make(chan struct{})
	defer close(done)

	logger := log.New(ioutil.Discard, "", 0)
	ts := httptest.NewServer(IndexHandler(logger, "testdata", done, testTempl))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 500, res.StatusCode, "got wrong response")
}
