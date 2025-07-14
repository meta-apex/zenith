package zcfg

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches file changes for hot reload
type FileWatcher struct {
	watcher  *fsnotify.Watcher
	filePath string
	config   *Config
	stopCh   chan struct{}
	running  bool
	mu       sync.RWMutex
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(c *Config) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:  watcher,
		filePath: c.file,
		config:   c,
		stopCh:   make(chan struct{}),
		running:  false,
	}

	return fw, nil
}

// Start starts watching the file
func (fw *FileWatcher) Start() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.running {
		return nil
	}

	err := fw.watcher.Add(fw.filePath)
	if err != nil {
		return err
	}

	fw.running = true

	go fw.watchLoop()

	return nil
}

// Stop stops watching the file
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.running {
		return nil
	}

	fw.running = false
	close(fw.stopCh)

	return fw.watcher.Close()
}

// IsRunning returns whether the watcher is running
func (fw *FileWatcher) IsRunning() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.running
}

// watchLoop is the main watch loop
func (fw *FileWatcher) watchLoop() {
	// Debounce timer to avoid multiple rapid file changes
	var timer *time.Timer
	var timerMu sync.Mutex

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Only handle write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Debounce file changes
				timerMu.Lock()
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(100*time.Millisecond, func() {
					fw.reloadConfig()
				})
				timerMu.Unlock()
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err // TODO: Add proper logging

		case <-fw.stopCh:
			return
		}
	}
}

// reloadConfig reloads configuration from file
func (fw *FileWatcher) reloadConfig() {
	// Parse the updated config file
	newRawMap, err := parseConfigFile(fw.filePath)
	if err != nil {
		// TODO: Add proper error handling/logging
		return
	}

	// Update the config
	if err := fw.config.Update(newRawMap); err != nil {
		// TODO: Add proper error handling/logging
		return
	}
}
