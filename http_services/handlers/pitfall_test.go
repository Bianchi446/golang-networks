package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWriteHeader(t *testing.T) {
	// First handler function: writes "Bad request" and sends BadRequest status
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest) // WriteHeader before writing body content
	}

	// Create a new HTTP request and response recorder
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	w := httptest.NewRecorder()

	// Call the handler
	handler(w, r)

	// Log the response status
	t.Logf("Response status: %q", w.Result().Status)

	// Second handler function: same logic but fixed status code usage
	handler = func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest) // Fixed status code
	}

	// Create another request and recorder for the second handler
	r = httptest.NewRequest(http.MethodGet, "http://test", nil)
	w = httptest.NewRecorder()

	// Call the second handler
	handler(w, r)

	// Log the second response status
	t.Logf("Response status: %q", w.Result().Status)
}
