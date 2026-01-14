package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ProgressEntry represents an entry in progress.txt
type ProgressEntry struct {
	Timestamp string
	Title     string
	Details   []string
}

var (
	// Match timestamp header: [2024-01-15 14:32] Completed: Task description
	timestampPattern = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2})\]\s*(.+)$`)
	// Match detail lines: - some detail
	detailPattern = regexp.MustCompile(`^[\s]*-\s*(.+)$`)
)

// ParseProgress parses progress.txt and extracts entries
func ParseProgress(filename string) ([]ProgressEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []ProgressEntry
	var currentEntry *ProgressEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			if currentEntry != nil {
				entries = append(entries, *currentEntry)
				currentEntry = nil
			}
			continue
		}

		// Check for timestamp header
		if matches := timestampPattern.FindStringSubmatch(line); matches != nil {
			// Save previous entry if exists
			if currentEntry != nil {
				entries = append(entries, *currentEntry)
			}
			currentEntry = &ProgressEntry{
				Timestamp: matches[1],
				Title:     matches[2],
				Details:   []string{},
			}
			continue
		}

		// Check for detail line
		if currentEntry != nil {
			if matches := detailPattern.FindStringSubmatch(line); matches != nil {
				currentEntry.Details = append(currentEntry.Details, matches[1])
			}
		}
	}

	// Don't forget the last entry
	if currentEntry != nil {
		entries = append(entries, *currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
