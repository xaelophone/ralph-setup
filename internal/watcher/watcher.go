package watcher

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watch watches the specified files and sends changes to the channel
func Watch(files []string, changes chan<- string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Add files to watch
	for _, file := range files {
		// Watch the directory containing the file
		dir := filepath.Dir(file)
		if dir == "" {
			dir = "."
		}

		err = watcher.Add(dir)
		if err != nil {
			log.Printf("Warning: Could not watch %s: %v", file, err)
		}
	}

	// Create a set of files we're interested in
	watchedFiles := make(map[string]bool)
	for _, f := range files {
		abs, _ := filepath.Abs(f)
		watchedFiles[abs] = true
		// Also add the base name for relative matching
		watchedFiles[filepath.Base(f)] = true
	}

	// Watch for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Check if this file is one we're watching
			basename := filepath.Base(event.Name)
			absname, _ := filepath.Abs(event.Name)

			if watchedFiles[basename] || watchedFiles[absname] {
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create {
					changes <- basename
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// WatchWithCallback watches files and calls the callback on changes
func WatchWithCallback(files []string, callback func(string)) error {
	changes := make(chan string)
	go func() {
		for file := range changes {
			callback(file)
		}
	}()
	return Watch(files, changes)
}
