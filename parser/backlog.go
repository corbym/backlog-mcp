package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// backlogEntryRe matches a numbered backlog entry in either format:
	//   old: 1. **STORY-013** — title *(status)*
	//   new: 1. [STORY-013](path) ([EPIC-NNN](path)) — title *(status)*
	// Capture group 1: story ID (e.g. STORY-013)
	backlogEntryRe = regexp.MustCompile(`^\d+\.\s+(?:\*\*|\[)(STORY-\d+)(?:\*\*|\])`)

	// backlogStatusRe matches the inline status marker *(in-progress)*
	backlogStatusRe = regexp.MustCompile(`\*\(([\w-]+)\)\*`)

	// leadingNumRe matches the leading ordinal in a backlog entry for renumbering
	leadingNumRe = regexp.MustCompile(`^\d+\.`)
)

// BacklogEntry is a parsed line from backlog.md.
type BacklogEntry struct {
	Position int
	StoryID  string
	Status   string // empty if no inline marker
	Raw      string
}

// ParseBacklog reads backlog.md and returns all entries in order.
func ParseBacklog(root string) ([]BacklogEntry, error) {
	path := filepath.Join(root, "backlog.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading backlog: %w", err)
	}

	var entries []BacklogEntry
	for _, line := range strings.Split(string(data), "\n") {
		m := backlogEntryRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		e := BacklogEntry{
			StoryID: m[1],
			Raw:     line,
		}
		if sm := backlogStatusRe.FindStringSubmatch(line); sm != nil {
			e.Status = sm[1]
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// RemoveFromBacklog removes the entry for storyID from backlog.md and
// renumbers the remaining entries sequentially from 1.
// Returns an error only if the file cannot be read or written; missing storyID is silently ignored.
func RemoveFromBacklog(root, storyID string) error {
	path := filepath.Join(root, "backlog.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading backlog: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	filtered := make([]string, 0, len(lines))
	counter := 1

	for _, line := range lines {
		m := backlogEntryRe.FindStringSubmatch(line)
		if m == nil {
			// non-entry line: pass through unchanged
			filtered = append(filtered, line)
			continue
		}
		if m[1] == storyID {
			// drop this entry
			continue
		}
		// renumber: replace leading "N." with "counter." preserving the rest of the line
		newLine := leadingNumRe.ReplaceAllString(line, fmt.Sprintf("%d.", counter))
		filtered = append(filtered, newLine)
		counter++
	}

	return writeAtomic(path, []byte(strings.Join(filtered, "\n")))
}

// UpdateBacklogStatus updates the inline status marker for storyID in backlog.md.
// If the story has no existing marker, one is appended. If newStatus is empty the marker is removed.
func UpdateBacklogStatus(root, storyID, newStatus string) error {
	path := filepath.Join(root, "backlog.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading backlog: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	found := false

	for i, line := range lines {
		m := backlogEntryRe.FindStringSubmatch(line)
		if m == nil || m[1] != storyID {
			continue
		}
		found = true
		// Strip any existing status marker and trailing whitespace, then append new one
		base := strings.TrimRight(backlogStatusRe.ReplaceAllString(line, ""), " ")
		if newStatus != "" {
			lines[i] = base + " *(" + newStatus + ")*"
		} else {
			lines[i] = base
		}
		break
	}

	if !found {
		return fmt.Errorf("story %s not found in backlog", storyID)
	}

	return writeAtomic(path, []byte(strings.Join(lines, "\n")))
}
