package main

import (
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("config error", "err", err)
		os.Exit(1)
	}

	state, err := LoadState(cfg.StatePath)
	if err != nil {
		slog.Warn("state load failed, starting fresh", "err", err)
		state = &State{Files: make(map[string]FileState)}
	}

	client := NewClient(cfg.AnythingLLMURL, cfg.AnythingLLMAPIKey)
	syncer := NewSyncer(cfg, client, state)

	// Wait for AnythingLLM to be reachable before proceeding (up to 5 min).
	// retryUntilReady returns the actual workspace slug assigned by AnythingLLM,
	// which may differ from the configured name (e.g. "intels-a1b2c3d4").
	actualSlug, err := retryUntilReady(client, cfg.WorkspaceSlug, 5*time.Minute)
	if err != nil {
		slog.Error("AnythingLLM not reachable", "err", err)
		os.Exit(1)
	}
	if actualSlug != cfg.WorkspaceSlug {
		slog.Info("workspace slug differs from configured name, using actual slug",
			"configured", cfg.WorkspaceSlug, "actual", actualSlug)
		cfg.WorkspaceSlug = actualSlug
	}

	// Initial full sync.
	if err := syncer.FullSync(); err != nil {
		slog.Error("initial sync error", "err", err)
		// Don't exit — watcher + periodic resync will recover.
	}

	// Set up filesystem watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("watcher create failed", "err", err)
		os.Exit(1)
	}
	defer watcher.Close()

	if err := watcher.Add(cfg.IntelsDir); err != nil {
		slog.Error("watcher add dir failed", "dir", cfg.IntelsDir, "err", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(cfg.SyncInterval)
	defer ticker.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("kb-maintainer running",
		"intelsDir", cfg.IntelsDir,
		"workspace", cfg.WorkspaceSlug,
		"syncInterval", cfg.SyncInterval.String(),
	)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			handleFSEvent(syncer, event)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("watcher error", "err", err)

		case <-ticker.C:
			if err := syncer.FullSync(); err != nil {
				slog.Error("periodic sync error", "err", err)
			}

		case <-sigs:
			slog.Info("shutting down")
			return
		}
	}
}

// handleFSEvent processes a single filesystem event from fsnotify.
func handleFSEvent(syncer *Syncer, event fsnotify.Event) {
	name := filepath.Base(event.Name)
	if !strings.HasSuffix(name, ".md") {
		return
	}
	switch {
	case event.Has(fsnotify.Create) || event.Has(fsnotify.Write):
		slog.Info("file event", "op", event.Op.String(), "file", name)
		// Small debounce: give the writer a moment to finish.
		time.Sleep(200 * time.Millisecond)
		if err := syncer.SyncFile(name); err != nil {
			slog.Error("sync failed", "file", name, "err", err)
		}
	case event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename):
		slog.Info("file removed", "file", name)
		if err := syncer.DeleteFile(name); err != nil {
			slog.Error("delete failed", "file", name, "err", err)
		}
	}
}

// retryUntilReady calls EnsureWorkspace with exponential backoff until it
// succeeds or deadline is exceeded. It returns the actual workspace slug.
func retryUntilReady(client *Client, workspaceName string, deadline time.Duration) (string, error) {
	start := time.Now()
	backoff := 5 * time.Second
	for {
		slug, err := client.EnsureWorkspace(workspaceName)
		if err == nil {
			return slug, nil
		}
		elapsed := time.Since(start)
		if elapsed+backoff > deadline {
			return "", err
		}
		slog.Warn("AnythingLLM not ready, retrying", "err", err, "retry_in", backoff.String())
		time.Sleep(backoff)
		if backoff < 60*time.Second {
			backoff *= 2
		}
	}
}
