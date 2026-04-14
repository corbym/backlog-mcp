package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupGroomRoot creates a minimal requirements dir with one epic and N story files.
func setupGroomRoot(t *testing.T, storyCount int) string {
	t.Helper()
	root := t.TempDir()

	// requirements-index.md
	indexContent := "# Requirements Index\n\n## EPIC-001: Groom Test Epic — `draft`\n\n| Story | Title | Status | Type |\n|-------|-------|--------|------|\n"
	for i := 1; i <= storyCount; i++ {
		indexContent += fmt.Sprintf("| [STORY-%03d](./epic-001-groom-test-epic/story-%03d.md) | Story %d title | draft | feature |\n", i, i, i)
	}
	os.WriteFile(filepath.Join(root, "requirements-index.md"), []byte(indexContent), 0o644)
	os.WriteFile(filepath.Join(root, "backlog.md"), []byte("# Backlog\n"), 0o644)

	epicDir := filepath.Join(root, "epic-001-groom-test-epic")
	os.MkdirAll(epicDir, 0o755)
	os.WriteFile(filepath.Join(epicDir, "epic-001.md"), []byte("# EPIC-001: Groom Test Epic\n\n## Goal\n\nTest.\n"), 0o644)

	for i := 1; i <= storyCount; i++ {
		os.WriteFile(
			filepath.Join(epicDir, fmt.Sprintf("story-%03d.md", i)),
			[]byte(fmt.Sprintf("# STORY-%03d: Story %d title\n\n## Goal\n\nDo it.\n\n## Acceptance criteria\n\n- [ ] Define acceptance criteria\n", i, i)),
			0o644,
		)
	}

	return root
}

func TestGroomEpic_AddsAllMissing(t *testing.T) {
	root := setupGroomRoot(t, 3)

	result, err := GroomEpic(root, "EPIC-001")
	if err != nil {
		t.Fatalf("GroomEpic: %v", err)
	}

	if len(result.Added) != 3 {
		t.Errorf("expected 3 added, got %v", result.Added)
	}
	if len(result.Removed) != 0 || len(result.Updated) != 0 {
		t.Errorf("expected no removes/updates, got removed=%v updated=%v", result.Removed, result.Updated)
	}

	data, _ := os.ReadFile(filepath.Join(root, "epic-001-groom-test-epic", "epic-001.md"))
	content := string(data)
	t.Logf("epic-001.md:\n%s", content)

	if !strings.Contains(content, "## Stories") {
		t.Error("expected ## Stories section")
	}
	for i := 1; i <= 3; i++ {
		entry := fmt.Sprintf("[STORY-%03d]", i)
		if !strings.Contains(content, entry) {
			t.Errorf("expected %s in epic, not found", entry)
		}
	}
}

func TestGroomEpic_NoOpWhenAlreadySynced(t *testing.T) {
	root := setupGroomRoot(t, 2)

	// First groom populates the section
	if _, err := GroomEpic(root, "EPIC-001"); err != nil {
		t.Fatalf("first groom: %v", err)
	}

	// Second groom should be a no-op
	result, err := GroomEpic(root, "EPIC-001")
	if err != nil {
		t.Fatalf("second groom: %v", err)
	}

	if len(result.Added) != 0 || len(result.Removed) != 0 || len(result.Updated) != 0 {
		t.Errorf("expected no-op on second groom, got %+v", result)
	}
	if len(result.Unchanged) != 2 {
		t.Errorf("expected 2 unchanged, got %v", result.Unchanged)
	}
}

func TestGroomEpic_RemovesStaleEntry(t *testing.T) {
	root := setupGroomRoot(t, 2)

	// Populate the section with both stories
	if _, err := GroomEpic(root, "EPIC-001"); err != nil {
		t.Fatalf("first groom: %v", err)
	}

	// Delete story-002.md from disk
	os.Remove(filepath.Join(root, "epic-001-groom-test-epic", "story-002.md"))

	result, err := GroomEpic(root, "EPIC-001")
	if err != nil {
		t.Fatalf("second groom: %v", err)
	}

	if len(result.Removed) != 1 || result.Removed[0] != "STORY-002" {
		t.Errorf("expected STORY-002 removed, got %v", result.Removed)
	}

	data, _ := os.ReadFile(filepath.Join(root, "epic-001-groom-test-epic", "epic-001.md"))
	if strings.Contains(string(data), "STORY-002") {
		t.Error("STORY-002 should have been removed from epic.md")
	}
}

func TestGroomEpic_RefreshesTitleAndStatus(t *testing.T) {
	root := setupGroomRoot(t, 1)

	// Populate section
	if _, err := GroomEpic(root, "EPIC-001"); err != nil {
		t.Fatalf("first groom: %v", err)
	}

	// Update index: change title and mark done in the story table row specifically
	indexPath := filepath.Join(root, "requirements-index.md")
	data, _ := os.ReadFile(indexPath)
	updated := strings.Replace(string(data), "| Story 1 title | draft | feature |", "| Updated title | done | feature |", 1)
	os.WriteFile(indexPath, []byte(updated), 0o644)

	result, err := GroomEpic(root, "EPIC-001")
	if err != nil {
		t.Fatalf("second groom: %v", err)
	}

	if len(result.Updated) != 1 || result.Updated[0] != "STORY-001" {
		t.Errorf("expected STORY-001 updated, got %v", result.Updated)
	}

	epicData, _ := os.ReadFile(filepath.Join(root, "epic-001-groom-test-epic", "epic-001.md"))
	content := string(epicData)
	t.Logf("epic-001.md:\n%s", content)

	if !strings.Contains(content, "- [x] [STORY-001]") {
		t.Error("expected story marked done with [x]")
	}
	if !strings.Contains(content, "Updated title") {
		t.Error("expected title refreshed to 'Updated title'")
	}
}

func TestGroomEpic_FallsBackToFileHeading(t *testing.T) {
	root := setupGroomRoot(t, 0) // no stories in index

	epicDir := filepath.Join(root, "epic-001-groom-test-epic")
	os.WriteFile(
		filepath.Join(epicDir, "story-001.md"),
		[]byte("# STORY-001: Orphan story title\n\n## Goal\n\nDo something.\n"),
		0o644,
	)

	result, err := GroomEpic(root, "EPIC-001")
	if err != nil {
		t.Fatalf("GroomEpic: %v", err)
	}

	if len(result.Added) != 1 {
		t.Errorf("expected 1 added, got %v", result.Added)
	}

	data, _ := os.ReadFile(filepath.Join(epicDir, "epic-001.md"))
	content := string(data)
	t.Logf("epic-001.md:\n%s", content)

	if !strings.Contains(content, "Orphan story title") {
		t.Errorf("expected title from story file heading, got:\n%s", content)
	}
}

func TestGroomEpic_EntriesInOrder(t *testing.T) {
	root := setupGroomRoot(t, 3)

	result, err := GroomEpic(root, "EPIC-001")
	if err != nil {
		t.Fatalf("GroomEpic: %v", err)
	}
	if len(result.Added) != 3 {
		t.Fatalf("expected 3 added, got %v", result.Added)
	}

	data, _ := os.ReadFile(filepath.Join(root, "epic-001-groom-test-epic", "epic-001.md"))
	content := string(data)

	pos1 := strings.Index(content, "STORY-001")
	pos2 := strings.Index(content, "STORY-002")
	pos3 := strings.Index(content, "STORY-003")

	if !(pos1 < pos2 && pos2 < pos3) {
		t.Errorf("expected ascending order: STORY-001 < STORY-002 < STORY-003, got positions %d %d %d", pos1, pos2, pos3)
	}
}
