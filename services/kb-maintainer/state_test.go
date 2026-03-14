package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadState(t *testing.T) {
	// 1. Loading a non-existent file
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "non-existent.json")
	state, err := LoadState(nonExistentPath)
	if err != nil {
		t.Fatalf("LoadState failed for non-existent file: %v", err)
	}
	if state.Files == nil || len(state.Files) != 0 {
		t.Errorf("Expected empty state for non-existent file, got %v", state)
	}

	// 2. Loading a valid state file
	validPath := filepath.Join(tmpDir, "valid.json")
	validData := `{
		"files": {
			"test.txt": {
				"hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"doc_location": "loc1",
				"uploaded_at": "2023-10-27T10:00:00Z"
			}
		}
	}`
	if err := os.WriteFile(validPath, []byte(validData), 0644); err != nil {
		t.Fatalf("Failed to write valid state file: %v", err)
	}
	state, err = LoadState(validPath)
	if err != nil {
		t.Fatalf("LoadState failed for valid file: %v", err)
	}
	if len(state.Files) != 1 || state.Files["test.txt"].Hash != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("Expected 1 file in state, got %v", state.Files)
	}

	// 3. Loading a corrupt state file
	corruptPath := filepath.Join(tmpDir, "corrupt.json")
	if err := os.WriteFile(corruptPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write corrupt state file: %v", err)
	}
	state, err = LoadState(corruptPath)
	if err != nil {
		t.Fatalf("LoadState should not return error for corrupt file: %v", err)
	}
	if state.Files == nil || len(state.Files) != 0 {
		t.Errorf("Expected empty state for corrupt file, got %v", state)
	}
}

func TestStateSave(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	uploadedAt, _ := time.Parse(time.RFC3339, "2023-10-27T10:00:00Z")

	state := &State{
		Files: map[string]FileState{
			"test.txt": {
				Hash:        "123",
				DocLocation: "loc",
				UploadedAt:  uploadedAt,
			},
		},
	}

	if err := state.Save(statePath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file was not created")
	}

	// Verify LoadState can read it back
	loadedState, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if len(loadedState.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(loadedState.Files))
	}

	f1 := state.Files["test.txt"]
	f2 := loadedState.Files["test.txt"]

	if f1.Hash != f2.Hash || f1.DocLocation != f2.DocLocation || !f1.UploadedAt.Equal(f2.UploadedAt) {
		t.Errorf("Loaded state doesn't match saved state.\nSaved: %+v\nLoaded: %+v", f1, f2)
	}
}

func TestHashContent(t *testing.T) {
	content := []byte("hello world")
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	actual := HashContent(content)
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}
