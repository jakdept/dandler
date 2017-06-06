package dandler

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
)

type canocialHostHandler struct {
	host    string
	port    string
	options int
	child   http.Handler
}

const (
	ForceHTTP      = 1 << iota // force http as the redirect target
	ForceHTTPS                 // force https as the redirect target
	ForceHost                  // force the given hostname as the redirect target
	ForcePort                  // force a given port for the redirect target
	ForceTemporary             // Use a 302 for the redirect
)

func CanocialHostHandler(host, port string, options int, childHandler http.Handler) http.Handler {
	return canocialHostHandler{host: host, port: port, options: options, child: childHandler}
}

func (h canocialHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.checkHostAndPort(r.Host) || h.checkScheme(r.TLS) {
		if h.options&ForceTemporary != 0 {
			http.Redirect(w, r, h.buildRedirect(*r.URL), http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, h.buildRedirect(*r.URL), http.StatusPermanentRedirect)
		}
	}
	h.child.ServeHTTP(w, r)
}

func (h canocialHostHandler) checkHostAndPort(url string) bool {
	if strings.Contains(url, ":") {
		chunks := strings.SplitN(url, ":", 2)
		if h.options&ForceHost != 0 && chunks[0] != h.host {
			return true
		}
		if h.options&ForcePort != 0 && chunks[1] != h.port {
			return true
		}
	} else {
		if h.options&ForceHost != 0 && url != h.host {
			return true
		}
	}
	return false
}

func (h canocialHostHandler) checkScheme(conn *tls.ConnectionState) bool {
	switch {
	case h.options&ForceHTTPS != 0 && conn == nil:
		return true
	case h.options&ForceHTTP != 0 && conn != nil:
		return true
	default:
		return false
	}
}

func (h canocialHostHandler) buildRedirect(url url.URL) string {
	// if host or port is forced, I have to modify the host header
	if h.options&(ForceHost|ForcePort) != 0 {
		var host, port string
		if strings.Contains(url.Host, ":") {
			chunks := strings.SplitN(url.Host, ":", 2)
			host, port = chunks[0], chunks[1]
		} else {
			host = url.Host
		}
		if h.options&ForceHost != 0 {
			host = h.host
		}
		if h.options&ForcePort != 0 {
			port = h.port
		}
		if port == "" {
			url.Host = host
		} else {
			url.Host = host + ":" + port
		}
	}

	// if forcing http, change it now
	if h.options&ForceHTTP != 0 {
		url.Scheme = "http"
	}

	// if forcing https, change it now
	if h.options&ForceHTTPS != 0 {
		url.Scheme = "https"
	}

	return url.String()
}
