package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// FileState tracks what has been uploaded to AnythingLLM.
type FileState struct {
	Hash        string    `json:"hash"`
	DocLocation string    `json:"doc_location"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// State is the persisted sync state, mapping filename → FileState.
type State struct {
	Files map[string]FileState `json:"files"`
}

// LoadState reads the state file; returns an empty state if the file does not exist.
func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &State{Files: make(map[string]FileState)}, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		// Corrupt state — start fresh rather than failing hard.
		return &State{Files: make(map[string]FileState)}, nil
	}
	if s.Files == nil {
		s.Files = make(map[string]FileState)
	}
	return &s, nil
}

// Save writes the state to disk atomically (write to temp, rename).
func (s *State) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// HashContent returns the MD5 hex digest of content.
func HashContent(content []byte) string {
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}
