package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// GroomResult describes what changed when reconciling an epic's Stories section.
type GroomResult struct {
	EpicID    string   `json:"epic_id"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Updated   []string `json:"updated"`
	Unchanged []string `json:"unchanged"`
}

// GroomEpic reconciles the ## Stories section of an epic.md file against the
// story files present on the filesystem and the metadata in requirements-index.md.
//
// It adds missing entries, removes entries for files that no longer exist,
// and refreshes titles and done/undone (- [x] / - [ ]) markers from the index.
// Entries are kept in ascending story-ID order.
func GroomEpic(root, epicID string) (GroomResult, error) {
	result := GroomResult{
		EpicID:    epicID,
		Added:     []string{},
		Removed:   []string{},
		Updated:   []string{},
		Unchanged: []string{},
	}

	epicDir, err := FindEpicDir(root, epicID)
	if err != nil {
		return result, err
	}

	epicNum, _ := parseIDNum(epicID)
	epicPath := filepath.Join(root, epicDir, fmt.Sprintf("epic-%03d.md", epicNum))

	data, err := os.ReadFile(epicPath)
	if err != nil {
		return result, fmt.Errorf("reading epic file: %w", err)
	}

	// ── 1. Index metadata for this epic ────────────────────────────────────────
	epics, err := ParseIndex(root)
	if err != nil {
		return result, err
	}

	type indexMeta struct{ title, status string }
	indexMap := map[string]indexMeta{}
	for _, e := range epics {
		if e.ID == epicID {
			for _, s := range e.Stories {
				indexMap[s.ID] = indexMeta{s.Title, s.Status}
			}
			break
		}
	}

	// ── 2. Story files on disk ──────────────────────────────────────────────────
	storyFileRe := regexp.MustCompile(`^story-(\d+)\.md$`)

	dirEntries, err := os.ReadDir(filepath.Join(root, epicDir))
	if err != nil {
		return result, fmt.Errorf("reading epic dir: %w", err)
	}

	type diskStory struct {
		id       string // STORY-NNN
		filename string // story-NNN.md
		num      int
	}
	var onDisk []diskStory
	for _, e := range dirEntries {
		if e.IsDir() {
			continue
		}
		m := storyFileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		onDisk = append(onDisk, diskStory{
			id:       fmt.Sprintf("STORY-%03d", n),
			filename: e.Name(),
			num:      n,
		})
	}
	sort.Slice(onDisk, func(i, j int) bool { return onDisk[i].num < onDisk[j].num })

	diskSet := map[string]string{} // storyID → filename
	for _, s := range onDisk {
		diskSet[s.id] = s.filename
	}

	// ── 3. Parse existing ## Stories section ───────────────────────────────────
	type existingEntry struct {
		checked bool
		title   string
	}
	storyEntryRe := regexp.MustCompile(`^- \[([ x])\] \[(STORY-\d+)\]\([^)]+\) — (.+)$`)

	lines := strings.Split(string(data), "\n")

	storiesIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Stories" {
			storiesIdx = i
			break
		}
	}

	sectionEnd := len(lines)
	if storiesIdx != -1 {
		for i := storiesIdx + 1; i < len(lines); i++ {
			if strings.HasPrefix(lines[i], "## ") {
				sectionEnd = i
				break
			}
		}
	}

	existingMap := map[string]existingEntry{}
	if storiesIdx != -1 {
		for _, line := range lines[storiesIdx+1 : sectionEnd] {
			m := storyEntryRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			existingMap[m[2]] = existingEntry{checked: m[1] == "x", title: m[3]}
		}
	}

	// ── 4. Reconcile ───────────────────────────────────────────────────────────
	type finalEntry struct {
		storyID  string
		filename string
		checked  bool
		title    string
	}
	var finalEntries []finalEntry

	for _, ds := range onDisk {
		meta, inIndex := indexMap[ds.id]
		existing, inSection := existingMap[ds.id]

		title := ""
		switch {
		case inIndex:
			title = meta.title
		case inSection:
			title = existing.title
		default:
			title = readStoryTitle(filepath.Join(root, epicDir, ds.filename), ds.id)
		}

		checked := inIndex && meta.status == "done"

		switch {
		case !inSection:
			result.Added = append(result.Added, ds.id)
		case existing.title != title || existing.checked != checked:
			result.Updated = append(result.Updated, ds.id)
		default:
			result.Unchanged = append(result.Unchanged, ds.id)
		}

		finalEntries = append(finalEntries, finalEntry{
			storyID:  ds.id,
			filename: ds.filename,
			checked:  checked,
			title:    title,
		})
	}

	// Entries in the section with no corresponding file on disk
	for id := range existingMap {
		if _, ok := diskSet[id]; !ok {
			result.Removed = append(result.Removed, id)
		}
	}
	sort.Strings(result.Removed)

	// Nothing changed — skip the write
	if len(result.Added) == 0 && len(result.Removed) == 0 && len(result.Updated) == 0 {
		return result, nil
	}

	// ── 5. Build the replacement Stories section ────────────────────────────────
	sectionLines := []string{"## Stories", ""}
	for _, fe := range finalEntries {
		check := " "
		if fe.checked {
			check = "x"
		}
		sectionLines = append(sectionLines, fmt.Sprintf("- [%s] [%s](%s) — %s", check, fe.storyID, fe.filename, fe.title))
	}
	sectionLines = append(sectionLines, "")

	// ── 6. Splice into the file ─────────────────────────────────────────────────
	var out []string
	if storiesIdx == -1 {
		trimmed := strings.TrimRight(string(data), "\n")
		out = strings.Split(trimmed+"\n\n"+strings.Join(sectionLines, "\n"), "\n")
	} else {
		out = append(out, lines[:storiesIdx]...)
		out = append(out, sectionLines...)
		out = append(out, lines[sectionEnd:]...)
	}

	return result, writeAtomic(epicPath, []byte(strings.Join(out, "\n")))
}

// readStoryTitle extracts the title from the first # heading in a story file.
// Falls back to the storyID if the file cannot be read or has no heading.
func readStoryTitle(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# ") {
			heading := strings.TrimPrefix(line, "# ")
			// "STORY-013: My title" → "My title"
			if idx := strings.Index(heading, ": "); idx != -1 {
				return strings.TrimSpace(heading[idx+2:])
			}
			return strings.TrimSpace(heading)
		}
	}
	return fallback
}
