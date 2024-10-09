package main

import (
	"encoding/json"
	"net/http"
)

// JSONResponse wraps a handler and ensures JSON responses
func JSONResponse(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a custom ResponseWriter
		rw := &responseWriter{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// If content type is not set, set it to application/json
		if rw.Header().Get("Content-Type") == "" {
			rw.Header().Set("Content-Type", "application/json")
		}

		// If no status has been written, write StatusOK
		if rw.status == 0 {
			rw.WriteHeader(http.StatusOK)
		}
	}
}

// responseWriter is a custom ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondError sends a JSON error response
func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}
