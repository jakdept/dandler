package dandler

import "net/http"

// MethodHandler splits requests to the Handler across the Method map.
type MethodHandler map[string]http.Handler

// ServeHTTP makes MethodHandler a http.Handler.
func (h MethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := h[r.Method]; ok {
		handler.ServeHTTP(w, r)
		return
	}
	http.Error(w, "unconfigured method", http.StatusInternalServerError)
}
