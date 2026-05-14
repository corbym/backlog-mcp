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

// ReorderBacklog rewrites backlog.md so entries appear in the order given by
// orderedIDs. IDs in the list that are not currently in the backlog are
// returned in notFound (not a hard error). Backlog entries absent from
// orderedIDs are appended at the end, preserving their relative order, so no
// entries are ever silently dropped. Renumbers all entries from 1.
func ReorderBacklog(root string, orderedIDs []string) (placed []string, notFound []string, err error) {
	path := filepath.Join(root, "backlog.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("reading backlog: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	// Build an id→raw-line map and record original entry order.
	entryLines := map[string]string{}
	var originalOrder []string
	for _, line := range lines {
		m := backlogEntryRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id := m[1]
		entryLines[id] = line
		originalOrder = append(originalOrder, id)
	}

	// Find IDs requested but not present in the backlog.
	for _, id := range orderedIDs {
		if _, ok := entryLines[id]; !ok {
			notFound = append(notFound, id)
		}
	}

	// Final order: requested (found) IDs first, then any leftovers in original order.
	seen := map[string]bool{}
	var finalOrder []string
	for _, id := range orderedIDs {
		if _, ok := entryLines[id]; ok {
			finalOrder = append(finalOrder, id)
			seen[id] = true
		}
	}
	for _, id := range originalOrder {
		if !seen[id] {
			finalOrder = append(finalOrder, id)
		}
	}

	// Rebuild: walk original lines, substitute each entry slot with the next
	// ID from finalOrder (preserving non-entry lines like headers/blanks).
	counter := 1
	entryIdx := 0
	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		m := backlogEntryRe.FindStringSubmatch(line)
		if m == nil {
			newLines = append(newLines, line)
			continue
		}
		if entryIdx >= len(finalOrder) {
			continue
		}
		newID := finalOrder[entryIdx]
		entryIdx++
		newLine := leadingNumRe.ReplaceAllString(entryLines[newID], fmt.Sprintf("%d.", counter))
		newLines = append(newLines, newLine)
		placed = append(placed, newID)
		counter++
	}

	return placed, notFound, writeAtomic(path, []byte(strings.Join(newLines, "\n")))
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
