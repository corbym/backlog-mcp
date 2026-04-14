package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ── fixture ──────────────────────────────────────────────────────────────────

const fixtureIndex = `# Requirements Index

## EPIC-001: Combat System — ` + "`draft`" + `

| Story | Title | Status |
|-------|-------|--------|
| [STORY-001](./epic-001-combat-system/story-001.md) | Basic combat | draft |
| [STORY-002](./epic-001-combat-system/story-002.md) | Enemy AI | in-progress |

## EPIC-002: Inventory — ` + "`in-progress`" + `

| Story | Title | Status |
|-------|-------|--------|
| [STORY-003](./epic-002-inventory/story-003.md) | Item pickup | done |
| [STORY-004](./epic-002-inventory/story-004.md) | Loot drops | in-progress |
`

const fixtureBacklog = `# Backlog

1. **STORY-001** — Basic combat
2. **STORY-002** — Enemy AI *(in-progress)*
3. **STORY-004** — Loot drops *(in-progress)*
`

const fixtureStory001 = `# STORY-001: Basic combat

Initial implementation of the combat system.

## Acceptance criteria

- [ ] Define acceptance criteria
`

const fixtureStory002 = `# STORY-002: Enemy AI

Enemy pathfinding and attack patterns.

## Acceptance criteria

- [x] Enemy can path to player
- [x] Enemy attacks when adjacent
`

const fixtureStory003 = `# STORY-003: Item pickup

Player can pick up items from the ground.

## Acceptance criteria

- [x] Player can pick up items
`

// fixtureStory004 has partial AC (one checked, one unchecked) for testing incomplete_items gating.
const fixtureStory004 = `# STORY-004: Loot drops

Enemies drop loot on death.

## Acceptance criteria

- [x] Enemy drops loot on death
- [ ] Loot respects rarity weights
`

func newFixture(t *testing.T) (root string, s *server.MCPServer) {
	t.Helper()
	root = t.TempDir()

	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("requirements-index.md", fixtureIndex)
	write("backlog.md", fixtureBacklog)
	write("epic-001-combat-system/story-001.md", fixtureStory001)
	write("epic-001-combat-system/story-002.md", fixtureStory002)
	write("epic-002-inventory/story-003.md", fixtureStory003)
	write("epic-002-inventory/story-004.md", fixtureStory004)

	s = buildServer(&Config{StoriesRoot: root})
	return root, s
}

// ── helpers ───────────────────────────────────────────────────────────────────

func callTool(t *testing.T, s *server.MCPServer, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	tool := s.GetTool(name)
	if tool == nil {
		t.Fatalf("tool %q not registered", name)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	result, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("tool %q handler error: %v", name, err)
	}
	return result
}

func resultText(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

// assertOK fails if the result is an error result.
func assertOK(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()
	if result.IsError {
		t.Fatalf("expected success result, got error: %s", resultText(result))
	}
}

// assertError fails if the result is not an error result.
func assertError(t *testing.T, result *mcp.CallToolResult, wantSubstring string) {
	t.Helper()
	if !result.IsError {
		t.Fatalf("expected error result, got success: %s", resultText(result))
	}
	if wantSubstring != "" && !strings.Contains(resultText(result), wantSubstring) {
		t.Fatalf("error %q does not contain %q", resultText(result), wantSubstring)
	}
}

func unmarshalArray(t *testing.T, result *mcp.CallToolResult) []map[string]any {
	t.Helper()
	assertOK(t, result)
	var out []map[string]any
	if err := json.Unmarshal([]byte(resultText(result)), &out); err != nil {
		t.Fatalf("unmarshal array: %v\nraw: %s", err, resultText(result))
	}
	return out
}

func unmarshalObject(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	assertOK(t, result)
	var out map[string]any
	if err := json.Unmarshal([]byte(resultText(result)), &out); err != nil {
		t.Fatalf("unmarshal object: %v\nraw: %s", err, resultText(result))
	}
	return out
}

// ── list_stories ─────────────────────────────────────────────────────────────

func TestListStories_ReturnsAllStories(t *testing.T) {
	_, s := newFixture(t)
	rows := unmarshalArray(t, callTool(t, s, "list_stories", map[string]any{}))
	if len(rows) != 4 {
		t.Fatalf("expected 4 stories, got %d", len(rows))
	}
}

func TestListStories_FilterByEpic(t *testing.T) {
	_, s := newFixture(t)
	rows := unmarshalArray(t, callTool(t, s, "list_stories", map[string]any{
		"epic_id": "EPIC-001",
	}))
	if len(rows) != 2 {
		t.Fatalf("expected 2 stories for EPIC-001, got %d", len(rows))
	}
	for _, r := range rows {
		if r["epic_id"] != "EPIC-001" {
			t.Errorf("unexpected epic_id %q", r["epic_id"])
		}
	}
}

func TestListStories_FilterByStatus(t *testing.T) {
	_, s := newFixture(t)
	rows := unmarshalArray(t, callTool(t, s, "list_stories", map[string]any{
		"status": "draft",
	}))
	if len(rows) != 1 {
		t.Fatalf("expected 1 draft story, got %d", len(rows))
	}
	if rows[0]["story_id"] != "STORY-001" {
		t.Errorf("expected STORY-001, got %q", rows[0]["story_id"])
	}
}

func TestListStories_FilterByEpicCaseInsensitive(t *testing.T) {
	_, s := newFixture(t)
	rows := unmarshalArray(t, callTool(t, s, "list_stories", map[string]any{
		"epic_id": "epic-001",
	}))
	if len(rows) != 2 {
		t.Fatalf("expected 2 stories for epic-001 (case-insensitive), got %d", len(rows))
	}
}

// ── get_story ─────────────────────────────────────────────────────────────────

func TestGetStory_ReturnsContentAndMetadata(t *testing.T) {
	_, s := newFixture(t)
	obj := unmarshalObject(t, callTool(t, s, "get_story", map[string]any{
		"story_id": "STORY-001",
	}))

	if obj["story_id"] != "STORY-001" {
		t.Errorf("story_id: got %q", obj["story_id"])
	}
	if obj["title"] != "Basic combat" {
		t.Errorf("title: got %q", obj["title"])
	}
	if obj["status"] != "draft" {
		t.Errorf("status: got %q", obj["status"])
	}
	if obj["epic_id"] != "EPIC-001" {
		t.Errorf("epic_id: got %q", obj["epic_id"])
	}
	content, _ := obj["content"].(string)
	if !strings.Contains(content, "Initial implementation") {
		t.Errorf("content missing expected text, got: %q", content)
	}
}

func TestGetStory_NotFound_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "get_story", map[string]any{
		"story_id": "STORY-999",
	})
	assertError(t, result, "STORY-999")
}

func TestGetStory_IDCaseInsensitive(t *testing.T) {
	_, s := newFixture(t)
	obj := unmarshalObject(t, callTool(t, s, "get_story", map[string]any{
		"story_id": "story-001",
	}))
	if obj["story_id"] != "STORY-001" {
		t.Errorf("expected STORY-001, got %q", obj["story_id"])
	}
}

// ── get_index_summary ─────────────────────────────────────────────────────────

func TestGetIndexSummary_ReturnsEpicsWithCounts(t *testing.T) {
	_, s := newFixture(t)
	rows := unmarshalArray(t, callTool(t, s, "get_index_summary", map[string]any{}))

	if len(rows) != 2 {
		t.Fatalf("expected 2 epics, got %d", len(rows))
	}

	var epic001 map[string]any
	for _, r := range rows {
		if r["epic_id"] == "EPIC-001" {
			epic001 = r
		}
	}
	if epic001 == nil {
		t.Fatal("EPIC-001 not found in summary")
	}

	counts, _ := epic001["counts"].(map[string]any)
	if counts["draft"] != float64(1) {
		t.Errorf("EPIC-001 draft count: got %v", counts["draft"])
	}
	if counts["in-progress"] != float64(1) {
		t.Errorf("EPIC-001 in-progress count: got %v", counts["in-progress"])
	}
}

// ── set_story_status ──────────────────────────────────────────────────────────

func TestSetStoryStatus_UpdatesIndexAndBacklog(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "set_story_status", map[string]any{
		"story_id": "STORY-001",
		"status":   "in-progress",
	}))

	if obj["old_status"] != "draft" {
		t.Errorf("old_status: got %q", obj["old_status"])
	}
	if obj["new_status"] != "in-progress" {
		t.Errorf("new_status: got %q", obj["new_status"])
	}
	if obj["backlog_updated"] != true {
		t.Errorf("expected backlog_updated=true, got %v", obj["backlog_updated"])
	}

	// Verify index file was actually updated on disk.
	index, _ := os.ReadFile(filepath.Join(root, "requirements-index.md"))
	if !strings.Contains(string(index), "in-progress") {
		t.Error("requirements-index.md not updated on disk")
	}

	// Verify backlog file was actually updated on disk.
	backlog, _ := os.ReadFile(filepath.Join(root, "backlog.md"))
	if !strings.Contains(string(backlog), "*(in-progress)*") {
		t.Error("backlog.md inline marker not updated on disk")
	}
}

func TestSetStoryStatus_Done_RedirectsToCompleteStory(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "set_story_status", map[string]any{
		"story_id": "STORY-001",
		"status":   "done",
	})
	assertError(t, result, "complete_story")
}

func TestSetStoryStatus_InvalidStatus_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "set_story_status", map[string]any{
		"story_id": "STORY-001",
		"status":   "wip",
	})
	assertError(t, result, "invalid status")
}

func TestSetStoryStatus_UnknownStory_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "set_story_status", map[string]any{
		"story_id": "STORY-999",
		"status":   "blocked",
	})
	assertError(t, result, "STORY-999")
}

func TestSetStoryStatus_MissingFromBacklog_ReturnsWarning(t *testing.T) {
	// STORY-003 is in the index (done) but not in backlog.md.
	// Setting it to in-progress should succeed but include backlog_warning.
	_, s := newFixture(t)
	obj := unmarshalObject(t, callTool(t, s, "set_story_status", map[string]any{
		"story_id": "STORY-003",
		"status":   "in-progress",
	}))
	if obj["backlog_updated"] != false {
		t.Errorf("expected backlog_updated=false, got %v", obj["backlog_updated"])
	}
	if obj["backlog_warning"] == nil || obj["backlog_warning"] == "" {
		t.Errorf("expected backlog_warning to be set, got %v", obj["backlog_warning"])
	}
}

// ── add_story_note ────────────────────────────────────────────────────────────

func TestAddStoryNote_AppendsNoteSection(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "add_story_note", map[string]any{
		"story_id": "STORY-001",
		"note":     "Verified combat animations look correct.",
	}))

	if obj["story_id"] != "STORY-001" {
		t.Errorf("story_id: got %q", obj["story_id"])
	}
	if obj["appended_at"] == "" {
		t.Error("expected appended_at timestamp")
	}

	content, _ := os.ReadFile(filepath.Join(root, "epic-001-combat-system", "story-001.md"))
	if !strings.Contains(string(content), "## Notes") {
		t.Error("## Notes section not added to story file")
	}
	if !strings.Contains(string(content), "Verified combat animations look correct.") {
		t.Error("note text not found in story file")
	}
	if !strings.Contains(string(content), "<!-- backlog-mcp:") {
		t.Error("backlog-mcp timestamp comment not found in story file")
	}
}

func TestAddStoryNote_SecondNote_AppendsWithoutDuplicatingSection(t *testing.T) {
	root, s := newFixture(t)

	callTool(t, s, "add_story_note", map[string]any{
		"story_id": "STORY-001",
		"note":     "First note.",
	})
	callTool(t, s, "add_story_note", map[string]any{
		"story_id": "STORY-001",
		"note":     "Second note.",
	})

	content, _ := os.ReadFile(filepath.Join(root, "epic-001-combat-system", "story-001.md"))
	count := strings.Count(string(content), "## Notes")
	if count != 1 {
		t.Errorf("expected exactly 1 ## Notes heading, got %d", count)
	}
	if !strings.Contains(string(content), "First note.") {
		t.Error("first note missing")
	}
	if !strings.Contains(string(content), "Second note.") {
		t.Error("second note missing")
	}
}

func TestAddStoryNote_UnknownStory_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "add_story_note", map[string]any{
		"story_id": "STORY-999",
		"note":     "Won't land.",
	})
	assertError(t, result, "STORY-999")
}

// ── create_epic ───────────────────────────────────────────────────────────────

func TestCreateEpic_CreatesDirectoryAndEpicFile(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "create_epic", map[string]any{
		"title":       "New Feature",
		"description": "Build the new feature.",
	}))

	epicID, _ := obj["epic_id"].(string)
	epicDir, _ := obj["path"].(string)

	if epicID == "" {
		t.Fatal("expected non-empty epic_id")
	}
	if epicDir == "" {
		t.Fatal("expected non-empty path")
	}

	// Directory should exist under root.
	info, err := os.Stat(filepath.Join(root, epicDir))
	if err != nil || !info.IsDir() {
		t.Errorf("epic directory %q not created: %v", epicDir, err)
	}

	// The epic .md file should exist inside the directory.
	n := epicDir[len("epic-") : len("epic-")+3] // e.g. "003"
	epicMD := filepath.Join(root, epicDir, "epic-"+n+".md")
	content, err := os.ReadFile(epicMD)
	if err != nil {
		t.Fatalf("epic.md not created at %s: %v", epicMD, err)
	}
	if !strings.Contains(string(content), "New Feature") {
		t.Errorf("epic.md missing title: %s", content)
	}
	if !strings.Contains(string(content), "Build the new feature.") {
		t.Errorf("epic.md missing description: %s", content)
	}

	// requirements-index.md should contain the new epic section.
	index, _ := os.ReadFile(filepath.Join(root, "requirements-index.md"))
	if !strings.Contains(string(index), epicID) {
		t.Errorf("requirements-index.md missing %s:\n%s", epicID, index)
	}
}

func TestCreateEpic_AssignsNextID(t *testing.T) {
	_, s := newFixture(t)

	// Fixture has EPIC-001 and EPIC-002 — next should be EPIC-003.
	obj := unmarshalObject(t, callTool(t, s, "create_epic", map[string]any{
		"title": "Third Epic",
	}))

	if obj["epic_id"] != "EPIC-003" {
		t.Errorf("expected EPIC-003, got %q", obj["epic_id"])
	}
}

func TestCreateEpic_SlugInPath(t *testing.T) {
	_, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "create_epic", map[string]any{
		"title": "My Cool Epic",
	}))

	path, _ := obj["path"].(string)
	if !strings.Contains(path, "my-cool-epic") {
		t.Errorf("expected slug in path, got %q", path)
	}
}

// ── create_story ──────────────────────────────────────────────────────────────

func TestCreateStory_CreatesFileAndRegisters(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "create_story", map[string]any{
		"epic_id":     "EPIC-001",
		"title":       "New story",
		"description": "Does something useful.",
	}))

	storyID, _ := obj["story_id"].(string)
	relPath, _ := obj["path"].(string)

	if storyID == "" {
		t.Fatal("expected non-empty story_id")
	}
	if relPath == "" {
		t.Fatal("expected non-empty path")
	}

	// Story file should exist.
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
	if err != nil {
		t.Fatalf("story file not created at %s: %v", relPath, err)
	}
	if !strings.Contains(string(content), "New story") {
		t.Errorf("story file missing title: %s", content)
	}
	if !strings.Contains(string(content), "Does something useful.") {
		t.Errorf("story file missing description: %s", content)
	}

	// requirements-index.md should contain the new story.
	index, _ := os.ReadFile(filepath.Join(root, "requirements-index.md"))
	if !strings.Contains(string(index), storyID) {
		t.Errorf("requirements-index.md missing %s:\n%s", storyID, index)
	}

	// backlog.md should contain the new story.
	backlog, _ := os.ReadFile(filepath.Join(root, "backlog.md"))
	if !strings.Contains(string(backlog), storyID) {
		t.Errorf("backlog.md missing %s:\n%s", storyID, backlog)
	}
}

func TestCreateStory_AssignsNextID(t *testing.T) {
	_, s := newFixture(t)

	// Fixture has STORY-001 through STORY-004 — next should be STORY-005.
	obj := unmarshalObject(t, callTool(t, s, "create_story", map[string]any{
		"epic_id": "EPIC-001",
		"title":   "Fifth story",
	}))

	if obj["story_id"] != "STORY-005" {
		t.Errorf("expected STORY-005, got %q", obj["story_id"])
	}
}

func TestCreateStory_EpicIDCaseInsensitive(t *testing.T) {
	_, s := newFixture(t)

	result := callTool(t, s, "create_story", map[string]any{
		"epic_id": "epic-001",
		"title":   "Case insensitive",
	})
	assertOK(t, result)
}

func TestCreateStory_UnknownEpic_ReturnsError(t *testing.T) {
	_, s := newFixture(t)

	result := callTool(t, s, "create_story", map[string]any{
		"epic_id": "EPIC-999",
		"title":   "Orphan story",
	})
	assertError(t, result, "EPIC-999")
}

// ── set_acceptance_criteria ───────────────────────────────────────────────────

func TestSetAcceptanceCriteria_ReplacesCriteriaSection(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "set_acceptance_criteria", map[string]any{
		"story_id": "STORY-001",
		"criteria": []any{"Combat starts when enemy is adjacent", "Player loses HP on hit"},
	}))

	if obj["story_id"] != "STORY-001" {
		t.Errorf("story_id: got %q", obj["story_id"])
	}
	if obj["criteria_count"] != float64(2) {
		t.Errorf("criteria_count: got %v", obj["criteria_count"])
	}
	if obj["path"] == "" {
		t.Error("expected non-empty path")
	}

	// Verify on disk: section replaced with new criteria.
	content, _ := os.ReadFile(filepath.Join(root, "epic-001-combat-system", "story-001.md"))
	body := string(content)
	if !strings.Contains(body, "## Acceptance criteria") {
		t.Error("## Acceptance criteria section missing")
	}
	if !strings.Contains(body, "- [ ] Combat starts when enemy is adjacent") {
		t.Error("first criterion missing")
	}
	if !strings.Contains(body, "- [ ] Player loses HP on hit") {
		t.Error("second criterion missing")
	}
	// Original story content should still be present.
	if !strings.Contains(body, "Initial implementation") {
		t.Error("original story content was clobbered")
	}
}

func TestSetAcceptanceCriteria_Idempotent(t *testing.T) {
	root, s := newFixture(t)

	callTool(t, s, "set_acceptance_criteria", map[string]any{
		"story_id": "STORY-001",
		"criteria": []any{"Old criterion"},
	})
	callTool(t, s, "set_acceptance_criteria", map[string]any{
		"story_id": "STORY-001",
		"criteria": []any{"New criterion A", "New criterion B"},
	})

	content, _ := os.ReadFile(filepath.Join(root, "epic-001-combat-system", "story-001.md"))
	body := string(content)
	if strings.Contains(body, "Old criterion") {
		t.Error("stale criterion from first call should have been replaced")
	}
	if !strings.Contains(body, "- [ ] New criterion A") {
		t.Error("New criterion A missing after second call")
	}
	if !strings.Contains(body, "- [ ] New criterion B") {
		t.Error("New criterion B missing after second call")
	}
	if strings.Count(body, "## Acceptance criteria") != 1 {
		t.Error("expected exactly one ## Acceptance criteria heading")
	}
}

func TestSetAcceptanceCriteria_EmptyCriteria_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "set_acceptance_criteria", map[string]any{
		"story_id": "STORY-001",
		"criteria": []any{},
	})
	assertError(t, result, "empty")
}

func TestSetAcceptanceCriteria_MissingCriteria_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "set_acceptance_criteria", map[string]any{
		"story_id": "STORY-001",
	})
	assertError(t, result, "criteria")
}

func TestSetAcceptanceCriteria_UnknownStory_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "set_acceptance_criteria", map[string]any{
		"story_id": "STORY-999",
		"criteria": []any{"Some criterion"},
	})
	assertError(t, result, "STORY-999")
}

// ── complete_story ─────────────────────────────────────────────────────────────

func TestCompleteStory_CompletesLifecycleAndAppendsSummary(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "complete_story", map[string]any{
		"story_id": "STORY-002",
		"summary":  "Implemented complete story flow and validated end-to-end.",
	}))

	if obj["story_id"] != "STORY-002" {
		t.Errorf("story_id: got %q", obj["story_id"])
	}
	if obj["completed_at"] == "" {
		t.Error("expected completed_at timestamp")
	}
	if obj["backlog_removed"] != true {
		t.Errorf("expected backlog_removed=true, got %v", obj["backlog_removed"])
	}

	index, _ := os.ReadFile(filepath.Join(root, "requirements-index.md"))
	if !strings.Contains(string(index), "| [STORY-002](./epic-001-combat-system/story-002.md) | Enemy AI | done |") {
		t.Error("requirements-index.md not updated to done for STORY-002")
	}

	backlog, _ := os.ReadFile(filepath.Join(root, "backlog.md"))
	if strings.Contains(string(backlog), "STORY-002") {
		t.Error("STORY-002 should have been removed from backlog.md")
	}

	story, _ := os.ReadFile(filepath.Join(root, "epic-001-combat-system", "story-002.md"))
	body := string(story)
	if !strings.Contains(body, "## Notes") {
		t.Error("## Notes section not added to story file")
	}
	if !strings.Contains(body, "Implemented complete story flow and validated end-to-end.") {
		t.Error("summary text not found in story file")
	}
	if !strings.Contains(body, "<!-- backlog-mcp:") {
		t.Error("backlog-mcp timestamp comment not found in story file")
	}
}

func TestCompleteStory_MissingSummary_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "complete_story", map[string]any{
		"story_id": "STORY-002",
	})
	assertError(t, result, "summary")
}

func TestCompleteStory_UnknownStory_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "complete_story", map[string]any{
		"story_id": "STORY-999",
		"summary":  "Done.",
	})
	assertError(t, result, "STORY-999")
}

func TestCompleteStory_AlreadyDone_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	result := callTool(t, s, "complete_story", map[string]any{
		"story_id": "STORY-003",
		"summary":  "Done again.",
	})
	assertError(t, result, "already done")
}

func TestCompleteStory_PlaceholderAC_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	// STORY-001 has only the placeholder "- [ ] Define acceptance criteria".
	result := callTool(t, s, "complete_story", map[string]any{
		"story_id": "STORY-001",
		"summary":  "Done.",
	})
	assertError(t, result, "set_acceptance_criteria")
}

func TestCompleteStory_AllChecked_Succeeds(t *testing.T) {
	_, s := newFixture(t)
	// STORY-002 has all criteria checked — no incomplete_items needed.
	obj := unmarshalObject(t, callTool(t, s, "complete_story", map[string]any{
		"story_id": "STORY-002",
		"summary":  "All criteria met.",
	}))
	if obj["story_id"] != "STORY-002" {
		t.Errorf("story_id: got %q", obj["story_id"])
	}
	if obj["completed_at"] == "" {
		t.Error("expected completed_at timestamp")
	}
}

func TestCompleteStory_SomeUnchecked_MissingIncompleteItems_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	// STORY-004 has one unchecked criterion — must provide incomplete_items.
	result := callTool(t, s, "complete_story", map[string]any{
		"story_id": "STORY-004",
		"summary":  "Partially done.",
	})
	assertError(t, result, "incomplete_items")
}

func TestCompleteStory_SomeUnchecked_WithIncompleteItems_Succeeds(t *testing.T) {
	root, s := newFixture(t)
	// STORY-004 has one unchecked criterion — provide one explanation.
	obj := unmarshalObject(t, callTool(t, s, "complete_story", map[string]any{
		"story_id":         "STORY-004",
		"summary":          "Shipped without rarity weights.",
		"incomplete_items": []any{"Deferred to STORY-010 — rarity system not yet designed"},
	}))
	if obj["story_id"] != "STORY-004" {
		t.Errorf("story_id: got %q", obj["story_id"])
	}

	// Note must contain both the summary and the incomplete criterion explanation.
	story, _ := os.ReadFile(filepath.Join(root, "epic-002-inventory", "story-004.md"))
	body := string(story)
	if !strings.Contains(body, "Shipped without rarity weights.") {
		t.Error("summary missing from note")
	}
	if !strings.Contains(body, "Incomplete criteria:") {
		t.Error("incomplete criteria section missing from note")
	}
	if !strings.Contains(body, "Loot respects rarity weights") {
		t.Error("unchecked criterion text missing from note")
	}
	if !strings.Contains(body, "Deferred to STORY-010") {
		t.Error("incomplete_items explanation missing from note")
	}
}

func TestCompleteStory_SomeUnchecked_CountMismatch_ReturnsError(t *testing.T) {
	_, s := newFixture(t)
	// STORY-004 has 1 unchecked criterion but we supply 2 explanations.
	result := callTool(t, s, "complete_story", map[string]any{
		"story_id":         "STORY-004",
		"summary":          "Done.",
		"incomplete_items": []any{"Explanation one", "Explanation two"},
	})
	assertError(t, result, "2 entries but there are 1 unchecked")
}

// ── check_acceptance_criterion ────────────────────────────────────────────────

func TestCheckAcceptanceCriterion_ByIndex_FlipsUnchecked(t *testing.T) {
	root, s := newFixture(t)

	// STORY-004: index 0 = checked, index 1 = unchecked "Loot respects rarity weights"
	obj := unmarshalObject(t, callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id":        "STORY-004",
		"criterion_index": float64(1),
	}))

	if obj["story_id"] != "STORY-004" {
		t.Errorf("story_id: got %q", obj["story_id"])
	}
	if obj["criterion"] != "Loot respects rarity weights" {
		t.Errorf("criterion: got %q", obj["criterion"])
	}
	if obj["checked"] != true {
		t.Errorf("checked: got %v", obj["checked"])
	}

	content, _ := os.ReadFile(filepath.Join(root, "epic-002-inventory", "story-004.md"))
	if !strings.Contains(string(content), "- [x] Loot respects rarity weights") {
		t.Error("criterion not flipped to [x] on disk")
	}
}

func TestCheckAcceptanceCriterion_ByText_FlipsUnchecked(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id":       "STORY-004",
		"criterion_text": "Loot respects rarity weights",
	}))

	if obj["criterion"] != "Loot respects rarity weights" {
		t.Errorf("criterion: got %q", obj["criterion"])
	}

	content, _ := os.ReadFile(filepath.Join(root, "epic-002-inventory", "story-004.md"))
	if !strings.Contains(string(content), "- [x] Loot respects rarity weights") {
		t.Error("criterion not flipped to [x] on disk")
	}
}

func TestCheckAcceptanceCriterion_ByText_CaseInsensitive(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id":       "STORY-004",
		"criterion_text": "loot RESPECTS rarity weights",
	}))

	if obj["criterion"] != "Loot respects rarity weights" {
		t.Errorf("criterion: got %q", obj["criterion"])
	}

	content, _ := os.ReadFile(filepath.Join(root, "epic-002-inventory", "story-004.md"))
	if !strings.Contains(string(content), "- [x] Loot respects rarity weights") {
		t.Error("criterion not flipped to [x] on disk")
	}
}

func TestCheckAcceptanceCriterion_AlreadyChecked_ReturnsError(t *testing.T) {
	_, s := newFixture(t)

	// STORY-004 index 0 is already [x]
	result := callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id":        "STORY-004",
		"criterion_index": float64(0),
	})
	assertError(t, result, "already checked")
}

func TestCheckAcceptanceCriterion_IndexOutOfRange_ReturnsError(t *testing.T) {
	_, s := newFixture(t)

	result := callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id":        "STORY-004",
		"criterion_index": float64(99),
	})
	assertError(t, result, "out of range")
}

func TestCheckAcceptanceCriterion_TextNotFound_ReturnsError(t *testing.T) {
	_, s := newFixture(t)

	result := callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id":       "STORY-004",
		"criterion_text": "this criterion does not exist",
	})
	assertError(t, result, "not found")
}

func TestCheckAcceptanceCriterion_UnknownStory_ReturnsError(t *testing.T) {
	_, s := newFixture(t)

	result := callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id":        "STORY-999",
		"criterion_index": float64(0),
	})
	assertError(t, result, "STORY-999")
}

func TestCheckAcceptanceCriterion_NeitherProvided_ReturnsError(t *testing.T) {
	_, s := newFixture(t)

	result := callTool(t, s, "check_acceptance_criterion", map[string]any{
		"story_id": "STORY-004",
	})
	assertError(t, result, "criterion_index or criterion_text")
}

// ── concurrency ───────────────────────────────────────────────────────────────

// TestConcurrentSetStoryStatus_Serialises fires 10 goroutines simultaneously,
// each transitioning a story status, and asserts the index file is valid
// (parseable, non-empty) after all writes complete — i.e. no interleaved
// corruption occurred.
func TestConcurrentSetStoryStatus_Serialises(t *testing.T) {
	root, s := newFixture(t)

	const workers = 10
	var wg sync.WaitGroup
	errs := make(chan string, workers)

	statuses := []string{"in-progress", "draft", "in-progress", "draft", "in-progress",
		"draft", "in-progress", "draft", "in-progress", "draft"}

	for i := range workers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			result := callTool(t, s, "set_story_status", map[string]any{
				"story_id": "STORY-001",
				"status":   statuses[idx],
			})
			if result.IsError {
				errs <- resultText(result)
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for e := range errs {
		t.Errorf("concurrent worker got error: %s", e)
	}

	// The index file must still be parseable and contain the story entry.
	data, err := os.ReadFile(filepath.Join(root, "requirements-index.md"))
	if err != nil {
		t.Fatalf("reading index after concurrent writes: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "STORY-001") {
		t.Error("requirements-index.md is corrupted: STORY-001 entry missing")
	}
	if !strings.Contains(content, "EPIC-001") {
		t.Error("requirements-index.md is corrupted: EPIC-001 section missing")
	}
}
