package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Syncer keeps the intels directory in sync with an AnythingLLM workspace.
type Syncer struct {
	cfg    *Config
	client *Client
	state  *State
}

// NewSyncer creates a Syncer.
func NewSyncer(cfg *Config, client *Client, state *State) *Syncer {
	return &Syncer{cfg: cfg, client: client, state: state}
}

// SyncFile uploads filename (relative to IntelsDir) if its content has changed
// since the last sync, then adds it to the workspace.
func (s *Syncer) SyncFile(filename string) error {
	fullPath := filepath.Join(s.cfg.IntelsDir, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filename, err)
	}

	hash := HashContent(content)
	if existing, ok := s.state.Files[filename]; ok && existing.Hash == hash {
		slog.Debug("unchanged, skip", "file", filename)
		return nil
	}

	slog.Info("uploading", "file", filename)
	location, err := s.client.UploadDocument(filename, content)
	if err != nil {
		return fmt.Errorf("upload %s: %w", filename, err)
	}

	if err := s.client.AddToWorkspace(s.cfg.WorkspaceSlug, []string{location}); err != nil {
		return fmt.Errorf("embed %s: %w", filename, err)
	}

	s.state.Files[filename] = FileState{
		Hash:        hash,
		DocLocation: location,
		UploadedAt:  time.Now(),
	}
	if err := s.state.Save(s.cfg.StatePath); err != nil {
		slog.Warn("state save failed", "err", err)
	}
	slog.Info("synced", "file", filename, "location", location)
	return nil
}

// DeleteFile removes filename from the workspace and clears it from state.
func (s *Syncer) DeleteFile(filename string) error {
	entry, ok := s.state.Files[filename]
	if !ok {
		return nil // not tracked, nothing to do
	}
	slog.Info("removing", "file", filename)
	if err := s.client.RemoveFromWorkspace(s.cfg.WorkspaceSlug, []string{entry.DocLocation}); err != nil {
		slog.Warn("remove from workspace failed", "file", filename, "err", err)
		// Continue — still clean up local state.
	}
	delete(s.state.Files, filename)
	if err := s.state.Save(s.cfg.StatePath); err != nil {
		slog.Warn("state save failed", "err", err)
	}
	return nil
}

// FullSync scans IntelsDir and syncs every .md file that is new or changed.
func (s *Syncer) FullSync() error {
	slog.Info("full sync started", "dir", s.cfg.IntelsDir)

	entries, err := os.ReadDir(s.cfg.IntelsDir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	seen := make(map[string]bool)
	var errs []string

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		seen[e.Name()] = true
		if err := s.SyncFile(e.Name()); err != nil {
			slog.Error("sync error", "file", e.Name(), "err", err)
			errs = append(errs, fmt.Sprintf("%s: %v", e.Name(), err))
		}
	}

	// Remove state entries for files that no longer exist.
	for name := range s.state.Files {
		if !seen[name] {
			if err := s.DeleteFile(name); err != nil {
				slog.Error("delete error", "file", name, "err", err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("sync completed with errors: %s", strings.Join(errs, "; "))
	}
	slog.Info("full sync complete")
	return nil
}
