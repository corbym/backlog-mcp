package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// CreateEpic assigns the next EPIC-NNN ID, creates the epic directory and
// epic markdown file, and appends its section header to requirements-index.md.
// Returns the epic ID and the directory name (relative to root).
func CreateEpic(root, title, description string) (epicID string, epicDir string, err error) {
	epics, err := ParseIndex(root)
	if err != nil {
		return "", "", err
	}

	n := nextEpicNum(epics)
	epicID = fmt.Sprintf("EPIC-%03d", n)
	epicDir = fmt.Sprintf("epic-%03d-%s", n, slugify(title))

	if err := os.MkdirAll(filepath.Join(root, epicDir), 0o755); err != nil {
		return "", "", fmt.Errorf("creating epic directory: %w", err)
	}

	epicMDPath := filepath.Join(root, epicDir, fmt.Sprintf("epic-%03d.md", n))
	if err := writeAtomic(epicMDPath, []byte(epicContent(epicID, title, description))); err != nil {
		return "", "", fmt.Errorf("writing epic file: %w", err)
	}

	if err := appendEpicToIndex(root, epicID, title); err != nil {
		return "", "", err
	}

	return epicID, epicDir, nil
}

// CreateStory assigns the next STORY-NNN ID, writes the story file, and
// updates requirements-index.md and backlog.md.
// Returns the story ID and the relative file path from root.
func CreateStory(root, epicID, title, description string) (storyID string, relPath string, err error) {
	epics, err := ParseIndex(root)
	if err != nil {
		return "", "", err
	}

	epicFound := false
	for _, e := range epics {
		if e.ID == epicID {
			epicFound = true
			break
		}
	}
	if !epicFound {
		return "", "", fmt.Errorf("epic %s not found in index", epicID)
	}

	n := nextStoryNum(epics)
	storyID = fmt.Sprintf("STORY-%03d", n)

	epicDir, err := FindEpicDir(root, epicID)
	if err != nil {
		return "", "", err
	}

	relPath = fmt.Sprintf("%s/story-%03d.md", epicDir, n)
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))

	if err := writeAtomic(fullPath, []byte(storyContent(storyID, title, description))); err != nil {
		return "", "", fmt.Errorf("writing story file: %w", err)
	}

	if err := appendStoryToIndex(root, epicID, storyID, title, relPath); err != nil {
		return "", "", err
	}

	if err := appendStoryToBacklog(root, storyID, title); err != nil {
		return "", "", err
	}

	return storyID, relPath, nil
}

// FindEpicDir returns the directory name (relative to root) for an epic,
// by scanning for a directory named epic-NNN-* under root.
func FindEpicDir(root, epicID string) (string, error) {
	n, err := parseIDNum(epicID)
	if err != nil {
		return "", fmt.Errorf("invalid epic ID %q: %w", epicID, err)
	}
	prefix := fmt.Sprintf("epic-%03d", n)

	entries, err := os.ReadDir(root)
	if err != nil {
		return "", fmt.Errorf("reading requirements dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			return e.Name(), nil
		}
	}
	return "", fmt.Errorf("directory for %s not found under %s", epicID, root)
}

// ── internal helpers ─────────────────────────────────────────────────────────

func nextEpicNum(epics []Epic) int {
	max := 0
	for _, e := range epics {
		if n, err := parseIDNum(e.ID); err == nil && n > max {
			max = n
		}
	}
	return max + 1
}

func nextStoryNum(epics []Epic) int {
	max := 0
	for _, e := range epics {
		for _, s := range e.Stories {
			if n, err := parseIDNum(s.ID); err == nil && n > max {
				max = n
			}
		}
	}
	return max + 1
}

// parseIDNum extracts the numeric part of an ID like STORY-007 or EPIC-003.
var idNumRe = regexp.MustCompile(`-(\d+)$`)

func parseIDNum(id string) (int, error) {
	m := idNumRe.FindStringSubmatch(id)
	if m == nil {
		return 0, fmt.Errorf("no numeric suffix in %q", id)
	}
	return strconv.Atoi(m[1])
}

// slugify converts a title to a lowercase, hyphen-separated filesystem slug.
var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = strings.TrimRight(s[:40], "-")
	}
	return s
}

func epicContent(epicID, title, description string) string {
	goal := description
	if goal == "" {
		goal = "Describe what this epic should accomplish."
	}
	return fmt.Sprintf("# %s: %s\n\n## Goal\n\n%s\n", epicID, title, goal)
}

func storyContent(storyID, title, description string) string {
	goal := description
	if goal == "" {
		goal = "Describe what this story should accomplish."
	}
	return fmt.Sprintf("# %s: %s\n\n## Goal\n\n%s\n\n## Acceptance criteria\n\n- [ ] Define acceptance criteria\n", storyID, title, goal)
}

func appendEpicToIndex(root, epicID, title string) error {
	path := filepath.Join(root, "requirements-index.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading index: %w", err)
	}
	content := strings.TrimRight(string(data), "\n")
	section := fmt.Sprintf("\n\n## %s: %s — `draft`\n\n| Story | Title | Status |\n|-------|-------|--------|\n", epicID, title)
	return writeAtomic(path, []byte(content+section))
}

func appendStoryToIndex(root, epicID, storyID, title, relPath string) error {
	path := filepath.Join(root, "requirements-index.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading index: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	epicHeadingRe := regexp.MustCompile(`^## ` + regexp.QuoteMeta(epicID) + `:`)

	epicLine := -1
	for i, line := range lines {
		if epicHeadingRe.MatchString(line) {
			epicLine = i
			break
		}
	}
	if epicLine == -1 {
		return fmt.Errorf("epic %s not found in index", epicID)
	}

	// find the last table row in this epic's section
	lastTableRow := epicLine
	for i := epicLine + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			break
		}
		if strings.HasPrefix(lines[i], "|") {
			lastTableRow = i
		}
	}

	newRow := fmt.Sprintf("| [%s](./%s) | %s | draft |", storyID, relPath, title)
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:lastTableRow+1]...)
	out = append(out, newRow)
	out = append(out, lines[lastTableRow+1:]...)

	return writeAtomic(path, []byte(strings.Join(out, "\n")))
}

func appendStoryToBacklog(root, storyID, title string) error {
	entries, err := ParseBacklog(root)
	if err != nil {
		return err
	}
	nextNum := len(entries) + 1

	path := filepath.Join(root, "backlog.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading backlog: %w", err)
	}

	content := strings.TrimRight(string(data), "\n")
	return writeAtomic(path, []byte(fmt.Sprintf("%s\n%d. **%s** — %s\n", content, nextNum, storyID, title)))
}