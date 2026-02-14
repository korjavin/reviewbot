package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	wrapped := Logging(handler)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("expected log output, got empty")
	}

	for _, want := range []string{"GET", "/health", "200"} {
		if !bytes.Contains([]byte(logOutput), []byte(want)) {
			t.Errorf("log output %q missing %q", logOutput, want)
		}
	}
}

func TestLoggingMiddlewareCaptures404(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	wrapped := Logging(handler)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}

	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("404")) {
		t.Errorf("log output %q missing 404", logOutput)
	}
}

func TestResponseWriterDefaultsTo200(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't call WriteHeader explicitly — default should be 200.
		w.Write([]byte("ok"))
	})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	wrapped := Logging(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("200")) {
		t.Errorf("log output %q missing 200 for default status", logOutput)
	}
}
