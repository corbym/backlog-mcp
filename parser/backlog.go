package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// matches:  1. **STORY-047** — description *(status)*
	backlogEntryRe = regexp.MustCompile(`^(\d+\.\s+\*\*)(STORY-\d+)(\*\*\s*—\s*.+?)(\s*\*\([\w-]+\)\*)?(\s*)$`)

	// matches the inline status marker  *(in-progress)*
	backlogStatusRe = regexp.MustCompile(`\*\(([\w-]+)\)\*`)
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
			StoryID: m[2],
			Raw:     line,
		}
		if sm := backlogStatusRe.FindStringSubmatch(m[4]); sm != nil {
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
		if m[2] == storyID {
			// drop this entry
			continue
		}
		// renumber
		newLine := fmt.Sprintf("%d. **%s%s%s%s", counter, m[2], m[3], m[4], m[5])
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
		if m == nil || m[2] != storyID {
			continue
		}
		found = true
		// rebuild the line without the old status marker
		base := m[1] + m[2] + m[3]
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
