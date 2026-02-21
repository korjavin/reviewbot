package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

// Client is a thin HTTP wrapper around the AnythingLLM REST API v1.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a Client pointing at the given AnythingLLM instance.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.httpClient.Do(req)
}

// ---- Workspace ---------------------------------------------------------

type workspaceResponse struct {
	Workspace *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"workspace"`
}

// EnsureWorkspace creates the workspace if it does not already exist.
func (c *Client) EnsureWorkspace(slug string) error {
	resp, err := c.doRequest("GET", "/api/v1/workspace/"+slug, nil, "")
	if err != nil {
		return fmt.Errorf("get workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var wr workspaceResponse
		if json.NewDecoder(resp.Body).Decode(&wr) == nil && wr.Workspace != nil {
			return nil // exists
		}
	}

	// Create workspace.
	body, _ := json.Marshal(map[string]string{"name": slug})
	resp2, err := c.doRequest("POST", "/api/v1/workspace/new", bytes.NewReader(body), "application/json")
	if err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK && resp2.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp2.Body)
		return fmt.Errorf("create workspace: status %d: %s", resp2.StatusCode, b)
	}
	return nil
}

// ---- Document upload ----------------------------------------------------

type uploadResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Documents []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Location string `json:"location"`
	} `json:"documents"`
}

// UploadDocument uploads file content to the AnythingLLM document store.
// It returns the document location string (used to embed the doc in a workspace).
func (c *Client) UploadDocument(filename string, content []byte) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := fw.Write(content); err != nil {
		return "", fmt.Errorf("write content: %w", err)
	}
	mw.Close()

	resp, err := c.doRequest("POST", "/api/v1/document/upload", &buf, mw.FormDataContentType())
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload: status %d: %s", resp.StatusCode, respBody)
	}

	var ur uploadResponse
	if err := json.Unmarshal(respBody, &ur); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}
	if !ur.Success {
		return "", fmt.Errorf("upload failed: %s", ur.Error)
	}
	if len(ur.Documents) == 0 {
		return "", fmt.Errorf("upload returned no documents")
	}
	return ur.Documents[0].Location, nil
}

// ---- Workspace embedding ------------------------------------------------

// AddToWorkspace embeds the given document locations into the workspace.
func (c *Client) AddToWorkspace(slug string, locations []string) error {
	return c.updateEmbeddings(slug, locations, nil)
}

// RemoveFromWorkspace removes the given document locations from the workspace.
func (c *Client) RemoveFromWorkspace(slug string, locations []string) error {
	return c.updateEmbeddings(slug, nil, locations)
}

func (c *Client) updateEmbeddings(slug string, adds, deletes []string) error {
	if adds == nil {
		adds = []string{}
	}
	if deletes == nil {
		deletes = []string{}
	}
	body, _ := json.Marshal(map[string]interface{}{
		"adds":    adds,
		"deletes": deletes,
	})
	resp, err := c.doRequest("POST", "/api/v1/workspace/"+slug+"/update-embeddings",
		bytes.NewReader(body), "application/json")
	if err != nil {
		return fmt.Errorf("update-embeddings: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update-embeddings: status %d: %s", resp.StatusCode, b)
	}
	return nil
}
