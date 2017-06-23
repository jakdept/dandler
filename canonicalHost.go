package dandler

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
)

type canonicalHost struct {
	host    string
	port    string
	options int
	child   http.Handler
}

// These constants are to be used with the Canocial Host Handler.
const (
	ForceHTTP      = 1 << iota // force http as the redirect target
	ForceHTTPS                 // force https as the redirect target
	ForceHost                  // force the given hostname as the redirect target
	ForcePort                  // force a given port for the redirect target
	ForceTemporary             // Use a 302 for the redirect
)

// CanonicalHost returns a http.Handler that redirects to the canocial host
// based on certain options. 0 may be passed for options if so desired, or provided
// bits can be forced on the client with a redirect.
func CanonicalHost(url string, options int, childHandler http.Handler) http.Handler {
	h := canonicalHost{options: options, child: childHandler}
	h.host, h.port = h.splitHostPort(url)
	return h
}

func (h canonicalHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.checkHostAndPort(r.Host) || h.checkScheme(r.TLS) {
		if h.options&ForceTemporary != 0 {
			http.Redirect(w, r, h.buildRedirect(r), http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, h.buildRedirect(r), http.StatusPermanentRedirect)
		}
	}
	h.child.ServeHTTP(w, r)
}

func (h canonicalHost) splitHostPort(url string) (string, string) {
	if !strings.Contains(url, ":") {
		return url, ""
	}
	parts := strings.SplitN(url, ":", 2)
	return parts[0], parts[1]
}

func (h canonicalHost) checkHostAndPort(url string) bool {
	host, port := h.splitHostPort(url)
	if h.options&ForceHost != 0 && host != h.host {
		return true
	}
	if h.options&ForcePort != 0 && port != h.port {
		return true
	}
	return false
}

func (h canonicalHost) checkScheme(conn *tls.ConnectionState) bool {
	switch {
	case h.options&ForceHTTPS != 0 && conn == nil:
		return true
	case h.options&ForceHTTP != 0 && conn != nil:
		return true
	default:
		return false
	}
}

func (h canonicalHost) buildRedirect(r *http.Request) string {
	// if host or port is forced, I have to modify the host header
	var host, port, scheme string
	host, port = h.splitHostPort(r.URL.Host)
	if r.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	// if forcing certain options, change them now
	if h.options&ForceHost != 0 {
		host = h.host
	}
	if h.options&ForcePort != 0 {
		port = h.port
	}
	if h.options&ForceHTTP != 0 {
		scheme = "http"
	}
	if h.options&ForceHTTPS != 0 {
		scheme = "https"
	}

	if port == "" {
		return fmt.Sprintf("%s://%s/%s", scheme, host, r.RequestURI)
	}
	return fmt.Sprintf("%s://%s:%s/%s", scheme, host, port, r.RequestURI)
}
