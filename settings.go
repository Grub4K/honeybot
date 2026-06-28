package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/fswatcher/fswatcher"
)

var settings = &Settings{
	path: "config.json",
}

type config struct {
	Token    string   `json:"token"`
	Channels Channels `json:"channels"`
}

type Channels = map[string]Channel

type Channel struct {
	Soft         bool     `json:"soft"`
	Delete       Duration `json:"delete"`
	IgnoredRoles []string `json:"ignored_roles"`
}

type Settings struct {
	mu   sync.RWMutex
	data config

	path string
}

func (s *Settings) Watch(ctx context.Context) error {
	watcher, err := fswatcher.NewWatcher()
	if err != nil {
		return fmt.Errorf("settings: watcher: create: %w", err)
	}
	defer watcher.Close()

	ops := fswatcher.Create | fswatcher.Write | fswatcher.Remove | fswatcher.Rename
	if err := watcher.Add(s.path, ops); err != nil {
		return fmt.Errorf("settings: watcher: add file: %w", err)
	}

	slog.Info(
		"watching settings file for changes",
		slog.String("path", s.path),
	)

	var wait <-chan time.Time
	for {
		select {
		case <-wait:
			wait = nil
			if err := settings.Load(); err != nil {
				slog.Error(
					"failed to read settings file",
					slog.Any("error", err),
				)
			} else {
				slog.Info(
					"reloaded settings",
				)
			}

		case event := <-watcher.Events:
			slog.Debug(
				"filewatcher event",
				slog.Any("name", event.Name),
				slog.String("event", event.Op.String()),
			)
			wait = time.After(time.Second)

		case err := <-watcher.Errors:
			slog.Error(
				"error while watching file",
				slog.Any("error", err),
			)

		case <-ctx.Done():
			return watcher.Close()
		}
	}
}

func (s *Settings) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("settings: open: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&s.data); err != nil {
		return fmt.Errorf("settings: parse: %w", err)
	}

	return nil
}

func (s *Settings) Get() config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

var _ json.Unmarshaler = (*Duration)(nil)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	duration, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}

	*d = Duration{duration}
	return nil
}
