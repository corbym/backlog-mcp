package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Story represents a single story entry from the index.
type Story struct {
	ID        string // e.g. STORY-009
	Title     string
	Status    string
	EpicID    string // e.g. EPIC-003
	StoryType string // e.g. feature, bug, chore, spike
}

// Epic represents an epic heading from the index.
type Epic struct {
	ID      string // e.g. EPIC-003
	Title   string
	Status  string
	Stories []Story
}

var (
	epicHeadingRe = regexp.MustCompile(`^## (EPIC-\d+): (.+?) — ` + "`" + `(\w[\w-]*)` + "`")
	// storyRowRe captures: story ID, title, status, and an optional type column.
	// Rows without a type column (legacy format) default to "feature" in ParseIndex.
	storyRowRe = regexp.MustCompile(`^\|\s*\[([^\]]+)\]\([^)]+\)\s*\|\s*([^|]+?)\s*\|\s*(\w[\w-]*)\s*\|(?:\s*(\w[\w-]*)\s*\|)?`)
)

// ParseIndex reads requirements-index.md and returns all epics and their stories.
func ParseIndex(root string) ([]Epic, error) {
	path := filepath.Join(root, "requirements-index.md")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening index: %w", err)
	}
	defer f.Close()

	var epics []Epic
	var current *Epic

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := epicHeadingRe.FindStringSubmatch(line); m != nil {
			if current != nil {
				epics = append(epics, *current)
			}
			current = &Epic{
				ID:     m[1],
				Title:  m[2],
				Status: m[3],
			}
			continue
		}

		if current != nil {
			if m := storyRowRe.FindStringSubmatch(line); m != nil {
				storyType := "feature"
				if len(m) > 4 && strings.TrimSpace(m[4]) != "" {
					storyType = strings.TrimSpace(m[4])
				}
				current.Stories = append(current.Stories, Story{
					ID:        strings.TrimSpace(m[1]),
					Title:     strings.TrimSpace(m[2]),
					Status:    strings.TrimSpace(m[3]),
					EpicID:    current.ID,
					StoryType: storyType,
				})
			}
		}
	}
	if current != nil {
		epics = append(epics, *current)
	}

	return epics, scanner.Err()
}

// UpdateEpicStatus updates the status marker in the epic heading line in requirements-index.md.
// Returns the old status, or an error if the epic is not found.
func UpdateEpicStatus(root, epicID, newStatus string) (oldStatus string, err error) {
	path := filepath.Join(root, "requirements-index.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading index: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	headingPattern := regexp.MustCompile(
		`^(## ` + regexp.QuoteMeta(epicID) + `: .+ — ` + "`" + `)(\w[\w-]*)` + "`" + `(.*)$`,
	)

	found := false
	for i, line := range lines {
		if m := headingPattern.FindStringSubmatch(line); m != nil {
			oldStatus = m[2]
			lines[i] = m[1] + newStatus + "`" + m[3]
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("epic %s not found in index", epicID)
	}

	return oldStatus, writeAtomic(path, []byte(strings.Join(lines, "\n")))
}

// UpdateStoryStatus updates the status cell for storyID in requirements-index.md.
// Returns the old status, or an error if the story is not found.
func UpdateStoryStatus(root, storyID, newStatus string) (oldStatus string, err error) {
	path := filepath.Join(root, "requirements-index.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading index: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	// Match a table row containing [STORY-NNN] link
	linkPattern := regexp.MustCompile(
		`^(\|\s*\[` + regexp.QuoteMeta(storyID) + `\]\([^)]+\)\s*\|[^|]+\|)\s*(\w[\w-]*)\s*(\|.*)$`,
	)

	found := false
	for i, line := range lines {
		if m := linkPattern.FindStringSubmatch(line); m != nil {
			oldStatus = m[2]
			lines[i] = m[1] + " " + newStatus + " " + m[3]
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("story %s not found in index", storyID)
	}

	return oldStatus, writeAtomic(path, []byte(strings.Join(lines, "\n")))
}
