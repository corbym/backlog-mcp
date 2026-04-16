package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// storyPathPattern matches epic-NNN-slug/story-NNN.md
var storyPathPattern = regexp.MustCompile(`(?i)^epic-\d+[^/]*/story-(\d+)\.md$`)

var (
	storyStatusBoldRe  = regexp.MustCompile(`^(\*\*Status:\*\*\s*)(\w[\w-]*)\s*$`)
	storyStatusPlainRe = regexp.MustCompile(`^(Status:\s*)(\w[\w-]*)\s*$`)
)

// FindStoryPath scans the filesystem under root for a file matching story-NNN.md
// inside any epic-* directory. Returns the relative path from root.
func FindStoryPath(root, storyID string) (string, error) {
	num := extractNumber(storyID)
	if num == "" {
		return "", fmt.Errorf("invalid story ID %q", storyID)
	}
	target := fmt.Sprintf("story-%s.md", num)

	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if !storyPathPattern.MatchString(rel) {
			return nil
		}
		if strings.EqualFold(d.Name(), target) {
			found = rel
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("story %s not found under %s", storyID, root)
	}
	return found, nil
}

// ReadStory reads the full markdown content of a story file.
func ReadStory(root, relPath string) (string, error) {
	full := filepath.Join(root, filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return "", fmt.Errorf("reading story %q: %w", relPath, err)
	}
	return string(data), nil
}

// AppendNote appends a timestamped note to a story file under a ## Notes section.
// If a ## Notes section already exists, the note is appended after the last existing note.
// Otherwise a new ## Notes section is created at the end of the file.
func AppendNote(root, relPath, timestamp, note string) error {
	full := filepath.Join(root, filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return fmt.Errorf("reading story for note: %w", err)
	}

	content := string(data)
	noteBlock := fmt.Sprintf("\n<!-- backlog-mcp: %s -->\n%s\n", timestamp, note)

	if strings.Contains(content, "## Notes") {
		content = content + noteBlock
	} else {
		content = content + "\n## Notes\n" + noteBlock
	}

	return writeAtomic(full, []byte(content))
}

// ACItem represents a single acceptance criterion and its checked state.
type ACItem struct {
	Text    string
	Checked bool
}

// ParseAcceptanceCriteria reads the ## Acceptance criteria section of a story
// file and returns each checklist item with its checked state.
// Returns nil if the section is not found.
func ParseAcceptanceCriteria(root, relPath string) ([]ACItem, error) {
	full := filepath.Join(root, filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("reading story for acceptance criteria: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	acStart := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Acceptance criteria" {
			acStart = i
			break
		}
	}
	if acStart == -1 {
		return nil, nil
	}

	acEnd := len(lines)
	for i := acStart + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			acEnd = i
			break
		}
	}

	var items []ACItem
	for _, line := range lines[acStart+1 : acEnd] {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "- [x] "), strings.HasPrefix(trimmed, "- [X] "):
			items = append(items, ACItem{Text: trimmed[6:], Checked: true})
		case strings.HasPrefix(trimmed, "- [ ] "):
			items = append(items, ACItem{Text: trimmed[6:], Checked: false})
		}
	}
	return items, nil
}

// SetAcceptanceCriteria replaces the ## Acceptance criteria section of a story
// file with the provided list of criteria. Each item becomes a `- [ ] ...` line.
// The operation is idempotent: calling it again replaces the previous content.
// All other sections (Goal, Notes, etc.) are left unchanged.
func SetAcceptanceCriteria(root, relPath string, criteria []string) error {
	full := filepath.Join(root, filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return fmt.Errorf("reading story for acceptance criteria: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	// Locate the ## Acceptance criteria heading.
	acStart := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Acceptance criteria" {
			acStart = i
			break
		}
	}
	if acStart == -1 {
		return fmt.Errorf("no '## Acceptance criteria' section found in %s", relPath)
	}

	// Find where the section ends: next ## heading or end of file.
	acEnd := len(lines)
	for i := acStart + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			acEnd = i
			break
		}
	}

	// Build the replacement section.
	replacement := make([]string, 0, len(criteria)+2)
	replacement = append(replacement, "## Acceptance criteria", "")
	for _, c := range criteria {
		replacement = append(replacement, "- [ ] "+c)
	}
	replacement = append(replacement, "")

	// Stitch together: everything before AC + replacement + everything after.
	out := make([]string, 0, len(lines))
	out = append(out, lines[:acStart]...)
	out = append(out, replacement...)
	out = append(out, lines[acEnd:]...)

	return writeAtomic(full, []byte(strings.Join(out, "\n")))
}

// CheckAcceptanceCriterion flips a single acceptance criterion from - [ ] to - [x].
// Identify the target by criterionIndex (0-based, pass -1 to ignore) or criterionText
// (case-insensitive exact match, pass "" to ignore). Exactly one must be non-sentinel.
// Returns the criterion text on success. Errors if the criterion is already checked,
// not found, or the story file cannot be read.
func CheckAcceptanceCriterion(root, relPath string, criterionIndex int, criterionText string) (string, error) {
	full := filepath.Join(root, filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return "", fmt.Errorf("reading story: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	// Locate the ## Acceptance criteria section.
	acStart := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Acceptance criteria" {
			acStart = i
			break
		}
	}
	if acStart == -1 {
		return "", fmt.Errorf("no '## Acceptance criteria' section found in %s", relPath)
	}
	acEnd := len(lines)
	for i := acStart + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			acEnd = i
			break
		}
	}

	// Collect checklist items and remember their position in lines.
	type acLine struct {
		lineIdx int
		item    ACItem
	}
	var acLines []acLine
	for i, line := range lines[acStart+1 : acEnd] {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "- [x] "), strings.HasPrefix(trimmed, "- [X] "):
			acLines = append(acLines, acLine{lineIdx: acStart + 1 + i, item: ACItem{Text: trimmed[6:], Checked: true}})
		case strings.HasPrefix(trimmed, "- [ ] "):
			acLines = append(acLines, acLine{lineIdx: acStart + 1 + i, item: ACItem{Text: trimmed[6:], Checked: false}})
		}
	}

	// Resolve the target line.
	targetIdx := -1
	targetText := ""

	if criterionIndex >= 0 {
		if criterionIndex >= len(acLines) {
			return "", fmt.Errorf("criterion index %d out of range: story has %d criteria", criterionIndex, len(acLines))
		}
		if acLines[criterionIndex].item.Checked {
			return "", fmt.Errorf("criterion %d (%q) is already checked", criterionIndex, acLines[criterionIndex].item.Text)
		}
		targetIdx = acLines[criterionIndex].lineIdx
		targetText = acLines[criterionIndex].item.Text
	} else {
		for _, ac := range acLines {
			if strings.EqualFold(ac.item.Text, criterionText) {
				if ac.item.Checked {
					return "", fmt.Errorf("criterion %q is already checked", ac.item.Text)
				}
				targetIdx = ac.lineIdx
				targetText = ac.item.Text
				break
			}
		}
		if targetIdx == -1 {
			return "", fmt.Errorf("criterion %q not found", criterionText)
		}
	}

	// Flip - [ ] to - [x] on the target line.
	lines[targetIdx] = strings.Replace(lines[targetIdx], "- [ ] ", "- [x] ", 1)

	return targetText, writeAtomic(full, []byte(strings.Join(lines, "\n")))
}

// UpdateStoryStatusMetadata updates a status metadata line in a story markdown file
// when one exists. Supported formats are:
//
//	**Status:** in-progress
//	Status: in-progress
//
// Returns true when a status line was found and rewritten.
func UpdateStoryStatusMetadata(root, relPath, newStatus string) (bool, error) {
	full := filepath.Join(root, filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return false, fmt.Errorf("reading story for status metadata: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	updated := false

	for i, line := range lines {
		if m := storyStatusBoldRe.FindStringSubmatch(line); m != nil {
			lines[i] = m[1] + newStatus
			updated = true
			break
		}
		if m := storyStatusPlainRe.FindStringSubmatch(line); m != nil {
			lines[i] = m[1] + newStatus
			updated = true
			break
		}
	}

	if !updated {
		return false, nil
	}

	return true, writeAtomic(full, []byte(strings.Join(lines, "\n")))
}

// extractNumber returns the zero-padded numeric portion of a story/epic ID.
// "STORY-009" -> "009", "STORY-47" -> "047" is NOT done here; we match as-is from the ID.
// Actually we just strip leading zeros and reformat to match filesystem convention.
func extractNumber(id string) string {
	parts := strings.SplitN(strings.ToUpper(id), "-", 2)
	if len(parts) != 2 {
		return ""
	}
	num := strings.TrimLeft(parts[1], "0")
	if num == "" {
		num = "0"
	}
	// pad to 3 digits to match story-NNN.md convention
	return fmt.Sprintf("%03s", num)
}
