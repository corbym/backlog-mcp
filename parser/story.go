package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
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
	ID      string // AC-STORY-NNN-XXXXXXXX, empty if not yet assigned
	Text    string
	Checked bool
}

// acIDRe matches the ID prefix "AC-STORY-NNN-XXXXXXXX: " at the start of a
// criterion's text part (after the "- [ ] " or "- [x] " marker).
var acIDRe = regexp.MustCompile(`^(AC-[A-Z]+-\d+-[0-9a-fA-F]{8}): (.*)$`)

// newACID generates a fresh AC criterion ID for the given storyID.
// Format: AC-STORY-NNN-XXXXXXXX where XXXXXXXX is the first 8 hex chars of a
// random UUID v4.
func newACID(storyID string) string {
	return fmt.Sprintf("AC-%s-%s", storyID, uuid.New().String()[:8])
}

// storyIDFromRelPath derives "STORY-NNN" from a relative path such as
// "epic-001-combat-system/story-001.md". Returns "" if the path does not match.
func storyIDFromRelPath(relPath string) string {
	base := filepath.Base(filepath.FromSlash(relPath))
	if !strings.HasPrefix(base, "story-") || !strings.HasSuffix(base, ".md") {
		return ""
	}
	num := base[len("story-") : len(base)-len(".md")]
	return "STORY-" + num
}

// parseACText splits "AC-STORY-042-a3f9b2c1: User can log in" into (id, text).
// If no recognised ID prefix is present, returns ("", fullText).
func parseACText(fullText string) (id, text string) {
	if m := acIDRe.FindStringSubmatch(fullText); m != nil {
		return m[1], m[2]
	}
	return "", fullText
}

// dashNormalizer replaces em-dash, en-dash, horizontal bar, figure dash, and
// minus sign with a plain hyphen-minus so that text-based AC lookups are
// tolerant of Unicode dash variants that LLMs and editors may substitute.
var dashNormalizer = strings.NewReplacer(
	"—", "-", // em dash
	"–", "-", // en dash
	"―", "-", // horizontal bar
	"‒", "-", // figure dash
	"−", "-", // minus sign
	"‐", "-", // hyphen
	"‑", "-", // non-breaking hyphen
)

// normalizeACKey folds dash variants and collapses whitespace so that AC text
// comparisons are tolerant of minor Unicode differences between what was stored
// and what an agent passes as a lookup key.
func normalizeACKey(s string) string {
	return strings.ToLower(strings.TrimSpace(dashNormalizer.Replace(s)))
}

// assignMissingIDs adds AC IDs to any criterion lines in the AC section that do
// not already have one. lines is the full file split by newline; acStart and
// acEnd delimit the lines belonging to the AC section (exclusive of the heading
// at acStart). storyID is used to construct the ID prefix. The slice is mutated
// in place.
func assignMissingIDs(lines []string, acStart, acEnd int, storyID string) {
	if storyID == "" {
		return
	}
	for i, line := range lines[acStart+1 : acEnd] {
		trimmed := strings.TrimSpace(line)
		var fullText, prefix string
		switch {
		case strings.HasPrefix(trimmed, "- [x] "), strings.HasPrefix(trimmed, "- [X] "):
			fullText = trimmed[6:]
			prefix = trimmed[:6]
		case strings.HasPrefix(trimmed, "- [ ] "):
			fullText = trimmed[6:]
			prefix = "- [ ] "
		default:
			continue
		}
		id, text := parseACText(fullText)
		if id == "" {
			lines[acStart+1+i] = prefix + newACID(storyID) + ": " + text
		}
	}
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
			id, text := parseACText(trimmed[6:])
			items = append(items, ACItem{ID: id, Text: text, Checked: true})
		case strings.HasPrefix(trimmed, "- [ ] "):
			id, text := parseACText(trimmed[6:])
			items = append(items, ACItem{ID: id, Text: text, Checked: false})
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

	// Build the replacement section, preserving existing IDs and assigning new
	// ones to any criterion that does not already carry an ID.
	storyID := storyIDFromRelPath(relPath)

	// Map existing criterion text → ID so we can reuse IDs for unchanged criteria.
	existingIDs := make(map[string]string)
	for _, line := range lines[acStart+1 : acEnd] {
		trimmed := strings.TrimSpace(line)
		var fullText string
		switch {
		case strings.HasPrefix(trimmed, "- [x] "), strings.HasPrefix(trimmed, "- [X] "):
			fullText = trimmed[6:]
		case strings.HasPrefix(trimmed, "- [ ] "):
			fullText = trimmed[6:]
		default:
			continue
		}
		if id, text := parseACText(fullText); id != "" {
			existingIDs[text] = id
		}
	}

	replacement := make([]string, 0, len(criteria)+2)
	replacement = append(replacement, "## Acceptance criteria", "")
	for _, c := range criteria {
		// The caller may pass the criterion with or without an ID prefix.
		existingID, text := parseACText(c)
		if existingID == "" {
			// No ID in the input — look for a previously assigned one.
			text = c
			if id, ok := existingIDs[text]; ok {
				existingID = id
			} else if storyID != "" {
				existingID = newACID(storyID)
			}
		}
		if existingID != "" {
			replacement = append(replacement, "- [ ] "+existingID+": "+text)
		} else {
			replacement = append(replacement, "- [ ] "+text)
		}
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
			id, text := parseACText(trimmed[6:])
			acLines = append(acLines, acLine{lineIdx: acStart + 1 + i, item: ACItem{ID: id, Text: text, Checked: true}})
		case strings.HasPrefix(trimmed, "- [ ] "):
			id, text := parseACText(trimmed[6:])
			acLines = append(acLines, acLine{lineIdx: acStart + 1 + i, item: ACItem{ID: id, Text: text, Checked: false}})
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
			// Match by ID (exact, case-insensitive) or by text (normalised, case-insensitive).
			matchByID := ac.item.ID != "" && strings.EqualFold(ac.item.ID, criterionText)
			matchByText := normalizeACKey(ac.item.Text) == normalizeACKey(criterionText)
			if matchByID || matchByText {
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

	// Lazily assign IDs to any criteria that do not have one yet.
	assignMissingIDs(lines, acStart, acEnd, storyIDFromRelPath(relPath))

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

// PatchAcceptanceCriteria updates the checked state of individual acceptance
// criteria. Criteria may be identified by their AC ID (e.g. "AC-STORY-042-a3f9b2c1")
// or by exact text match; ID is preferred when both would match. Only criteria
// present in the updates map are modified; all others are left unchanged.
//
// If any key in updates does not match any criterion (by ID or text), the
// function returns the list of unmatched keys and leaves the file unchanged.
// On a successful write, IDs are lazily assigned to any criteria that did not
// already have one.
func PatchAcceptanceCriteria(root, relPath string, updates map[string]bool) (notFound []string, err error) {
	full := filepath.Join(root, filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("reading story: %w", err)
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
		return nil, fmt.Errorf("no '## Acceptance criteria' section found in %s", relPath)
	}
	acEnd := len(lines)
	for i := acStart + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			acEnd = i
			break
		}
	}

	// Build maps from criterion text and ID to line metadata.
	type acEntry struct {
		lineIdx int
		checked bool
	}
	acByText := make(map[string]acEntry)
	acByID := make(map[string]acEntry)
	for i, line := range lines[acStart+1 : acEnd] {
		trimmed := strings.TrimSpace(line)
		var fullText string
		var checked bool
		switch {
		case strings.HasPrefix(trimmed, "- [x] "), strings.HasPrefix(trimmed, "- [X] "):
			fullText = trimmed[6:]
			checked = true
		case strings.HasPrefix(trimmed, "- [ ] "):
			fullText = trimmed[6:]
			checked = false
		default:
			continue
		}
		id, text := parseACText(fullText)
		entry := acEntry{lineIdx: acStart + 1 + i, checked: checked}
		acByText[normalizeACKey(text)] = entry
		if id != "" {
			acByID[id] = entry
		}
	}

	// Validate that all requested criteria exist (by ID or by text).
	for key := range updates {
		_, byID := acByID[key]
		_, byText := acByText[normalizeACKey(key)]
		if !byID && !byText {
			notFound = append(notFound, key)
		}
	}
	if len(notFound) > 0 {
		return notFound, fmt.Errorf("criterion/criteria not found: %s", strings.Join(notFound, ", "))
	}

	// Apply updates, preferring ID match over text match.
	for key, wantChecked := range updates {
		var entry acEntry
		if e, ok := acByID[key]; ok {
			entry = e
		} else {
			entry = acByText[normalizeACKey(key)]
		}
		if wantChecked && !entry.checked {
			lines[entry.lineIdx] = strings.Replace(lines[entry.lineIdx], "- [ ] ", "- [x] ", 1)
		} else if !wantChecked && entry.checked {
			line := lines[entry.lineIdx]
			if len(line) >= 6 && strings.EqualFold(line[:6], "- [x] ") {
				lines[entry.lineIdx] = "- [ ] " + line[6:]
			}
		}
	}

	// Lazily assign IDs to any criteria that do not have one yet.
	assignMissingIDs(lines, acStart, acEnd, storyIDFromRelPath(relPath))

	return nil, writeAtomic(full, []byte(strings.Join(lines, "\n")))
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
