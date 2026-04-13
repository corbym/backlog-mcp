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
