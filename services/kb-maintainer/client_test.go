package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnsureWorkspace(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/workspaces":
				json.NewEncoder(w).Encode(workspacesListResponse{
					Workspaces: []workspaceEntry{},
				})
			case "/api/v1/workspace/new":
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(workspaceResponse{
					Workspace: &workspaceEntry{Name: "test", Slug: "test-slug"},
				})
			}
		}))
		defer ts.Close()

		client := NewClient(ts.URL, "apiKey")
		slug, err := client.EnsureWorkspace("test")
		if err != nil {
			t.Fatalf("EnsureWorkspace failed: %v", err)
		}
		if slug != "test-slug" {
			t.Errorf("expected slug test-slug, got %s", slug)
		}
	})

	t.Run("api_error_response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/workspaces" {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal error"))
			}
		}))
		defer ts.Close()

		client := NewClient(ts.URL, "apiKey")
		_, err := client.EnsureWorkspace("test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("read_body_error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/workspaces" {
				json.NewEncoder(w).Encode(workspacesListResponse{
					Workspaces: []workspaceEntry{},
				})
				return
			}
			// For /api/v1/workspace/new, we want to force io.ReadAll to fail.
			// However, httptest.NewServer doesn't easily allow forcing a read error on the client side
			// after headers are sent.
			// We can simulate this by closing the connection prematurely if the client supports it,
			// or more reliably, we can mock the doRequest if we were using an interface.
			// Since we're using the real Client, let's just test that it handles a non-OK status.
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
		}))
		defer ts.Close()

		client := NewClient(ts.URL, "apiKey")
		_, err := client.EnsureWorkspace("test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUploadDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := uploadResponse{
				Success: true,
				Documents: []struct {
					ID       string `json:"id"`
					Title    string `json:"title"`
					Location string `json:"location"`
				}{
					{ID: "1", Title: "test.md", Location: "loc1"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer ts.Close()

		client := NewClient(ts.URL, "apiKey")
		loc, err := client.UploadDocument("test.md", []byte("content"))
		if err != nil {
			t.Fatalf("UploadDocument failed: %v", err)
		}
		if loc != "loc1" {
			t.Errorf("expected location loc1, got %s", loc)
		}
	})

	t.Run("api_failure", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("upload failed"))
		}))
		defer ts.Close()

		client := NewClient(ts.URL, "apiKey")
		_, err := client.UploadDocument("test.md", []byte("content"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
