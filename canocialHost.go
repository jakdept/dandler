package dandler

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
)

type canocialHostHandler struct {
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

// CanocialHost returns a http.Handler that redirects to the canocial host
// based on certain options. 0 may be passed for options if so desired, or provided
// bits can be forced on the client with a redirect.
func CanocialHost(host, port string, options int, childHandler http.Handler) http.Handler {
	return canocialHostHandler{host: host, port: port, options: options, child: childHandler}
}

func (h canocialHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.checkHostAndPort(r.Host) || h.checkScheme(r.TLS) {
		if h.options&ForceTemporary != 0 {
			http.Redirect(w, r, h.buildRedirect(r), http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, h.buildRedirect(r), http.StatusPermanentRedirect)
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

func (h canocialHostHandler) buildRedirect(r *http.Request) string {
	// if host or port is forced, I have to modify the host header
	var host, port, scheme string
	if h.options&(ForceHost|ForcePort) != 0 {
		if strings.Contains(r.Host, ":") {
			chunks := strings.SplitN(r.Host, ":", 2)
			host, port = chunks[0], chunks[1]
		} else {
			host = r.Host
		}
	}

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
