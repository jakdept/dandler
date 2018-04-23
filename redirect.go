package dandler

import (
	"net/http"
	"path"
)

// RedirectURIHandler creates a handler that responds with a redirect to another
// domain. The URI is preserved, optionally with a prefix added. https is
// included or not, based on the value of TLS.
type RedirectURIHandler struct {
	TLS       bool
	Domain    string
	PrefixURI string
	Code      int
}

// ServeHTTP makes RedirectURIHandler a http.Handler.
func (rh *RedirectURIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := path.Join(rh.Domain, rh.PrefixURI, r.RequestURI)
	if rh.TLS {
		target = "https://" + target
	} else {
		target = "http://" + target
	}
	http.Redirect(w, r, target, rh.Code)
}
