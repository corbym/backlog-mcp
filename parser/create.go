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
// storyType should be one of: feature, bug, chore, spike.
// Returns the story ID and the relative file path from root.
func CreateStory(root, epicID, title, description, storyType string) (storyID string, relPath string, err error) {
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

	if err := writeAtomic(fullPath, []byte(storyContent(storyID, title, description, storyType))); err != nil {
		return "", "", fmt.Errorf("writing story file: %w", err)
	}

	if err := appendStoryToIndex(root, epicID, storyID, title, relPath, storyType); err != nil {
		return "", "", err
	}

	if err := appendStoryToBacklog(root, epicDir, epicID, storyID, relPath, title); err != nil {
		return "", "", err
	}

	if err := appendStoryToEpic(root, epicDir, epicID, storyID, relPath, title); err != nil {
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

func storyContent(storyID, title, description, storyType string) string {
	goal := description
	if goal == "" {
		goal = "Describe what this story should accomplish."
	}
	return fmt.Sprintf("# %s: %s\n\n**Type:** %s\n\n## Goal\n\n%s\n\n## Acceptance criteria\n\n- [ ] Define acceptance criteria\n", storyID, title, storyType, goal)
}

func appendEpicToIndex(root, epicID, title string) error {
	path := filepath.Join(root, "requirements-index.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading index: %w", err)
	}
	content := strings.TrimRight(string(data), "\n")
	section := fmt.Sprintf("\n\n## %s: %s — `draft`\n\n| Story | Title | Status | Type |\n|-------|-------|--------|------|\n", epicID, title)
	return writeAtomic(path, []byte(content+section))
}

func appendStoryToIndex(root, epicID, storyID, title, relPath, storyType string) error {
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

	newRow := fmt.Sprintf("| [%s](./%s) | %s | draft | %s |", storyID, relPath, title, storyType)
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:lastTableRow+1]...)
	out = append(out, newRow)
	out = append(out, lines[lastTableRow+1:]...)

	return writeAtomic(path, []byte(strings.Join(out, "\n")))
}

// appendStoryToBacklog adds a new story entry to the end of backlog.md with relative
// links to both the story file and the parent epic file.
func appendStoryToBacklog(root, epicDir, epicID, storyID, storyRelPath, title string) error {
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

	epicNum, _ := parseIDNum(epicID)
	epicFilePath := fmt.Sprintf("%s/epic-%03d.md", epicDir, epicNum)

	content := strings.TrimRight(string(data), "\n")
	entry := fmt.Sprintf("%d. [%s](%s) ([%s](%s)) — %s\n",
		nextNum, storyID, storyRelPath, epicID, epicFilePath, title)
	return writeAtomic(path, []byte(content+"\n"+entry))
}

// appendStoryToEpic adds a new entry to the ## Stories section of the parent epic.md.
// If the section does not exist yet it is appended to the file.
func appendStoryToEpic(root, epicDir, epicID, storyID, storyRelPath, title string) error {
	epicNum, _ := parseIDNum(epicID)
	epicPath := filepath.Join(root, epicDir, fmt.Sprintf("epic-%03d.md", epicNum))

	data, err := os.ReadFile(epicPath)
	if err != nil {
		return fmt.Errorf("reading epic file: %w", err)
	}

	// Link uses just the filename since epic.md and story.md share the same directory
	storyFilename := filepath.Base(filepath.FromSlash(storyRelPath))
	newEntry := fmt.Sprintf("- [ ] [%s](%s) — %s", storyID, storyFilename, title)

	lines := strings.Split(string(data), "\n")

	// Find the ## Stories section
	storiesIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Stories" {
			storiesIdx = i
			break
		}
	}

	if storiesIdx == -1 {
		// No section yet — append it
		trimmed := strings.TrimRight(string(data), "\n")
		return writeAtomic(epicPath, []byte(trimmed+"\n\n## Stories\n\n"+newEntry+"\n"))
	}

	// Find end of section: next ## heading or end of file
	sectionEnd := len(lines)
	for i := storiesIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			sectionEnd = i
			break
		}
	}

	// Insert after the last non-empty line in the section
	insertAt := storiesIdx + 1
	for i := storiesIdx + 1; i < sectionEnd; i++ {
		if strings.TrimSpace(lines[i]) != "" {
			insertAt = i + 1
		}
	}

	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:insertAt]...)
	out = append(out, newEntry)
	out = append(out, lines[insertAt:]...)
	return writeAtomic(epicPath, []byte(strings.Join(out, "\n")))
}

// MarkEpicStoryDone flips a story's entry in the parent epic.md from "- [ ]" to "- [x]".
// No-op if the entry is not found or already checked.
func MarkEpicStoryDone(root, epicID, storyID string) error {
	epicDir, err := FindEpicDir(root, epicID)
	if err != nil {
		return err
	}

	epicNum, _ := parseIDNum(epicID)
	epicPath := filepath.Join(root, epicDir, fmt.Sprintf("epic-%03d.md", epicNum))

	data, err := os.ReadFile(epicPath)
	if err != nil {
		return fmt.Errorf("reading epic file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	storyLinkRe := regexp.MustCompile(`^- \[ \] \[` + regexp.QuoteMeta(storyID) + `\]`)
	for i, line := range lines {
		if storyLinkRe.MatchString(line) {
			lines[i] = strings.Replace(line, "- [ ] ", "- [x] ", 1)
			break
		}
	}

	return writeAtomic(epicPath, []byte(strings.Join(lines, "\n")))
}