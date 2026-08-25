package gitcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// newTestServer creates a test HTTP server with the given handler and returns a client.
func newTestServer(handler http.Handler) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	return client, server
}

// jsonHandler returns an http.HandlerFunc that writes the given value as JSON.
func jsonHandler(statusCode int, v interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if v != nil {
			json.NewEncoder(w).Encode(v)
		}
	}
}

// jsonResponse returns a handler that responds with JSON for any method/path.
func jsonResponse(statusCode int, body interface{}) http.HandlerFunc {
	return jsonHandler(statusCode, body)
}

// methodRouter returns a handler that dispatches based on HTTP method.
//
//nolint:unused
func methodRouter(routes map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := routes[r.Method]; ok {
			handler(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// pathMethodRouter returns a handler that dispatches based on method+path.
//
//nolint:unused
func pathMethodRouter(routes map[string]map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if methods, ok := routes[r.URL.Path]; ok {
			if handler, ok := methods[r.Method]; ok {
				handler(w, r)
				return
			}
		}
		// Try prefix matching
		for path, methods := range routes {
			if len(r.URL.Path) >= len(path) && r.URL.Path[:len(path)] == path {
				if handler, ok := methods[r.Method]; ok {
					handler(w, r)
					return
				}
			}
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"404 Not Found"}`))
	}
}
