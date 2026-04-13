package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
`

const fixtureBacklog = `# Backlog

1. **STORY-001** — Basic combat
2. **STORY-002** — Enemy AI *(in-progress)*
`

const fixtureStory001 = `# STORY-001: Basic combat

Initial implementation of the combat system.
`

const fixtureStory002 = `# STORY-002: Enemy AI

Enemy pathfinding and attack patterns.
`

const fixtureStory003 = `# STORY-003: Item pickup

Player can pick up items from the ground.
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
	if len(rows) != 3 {
		t.Fatalf("expected 3 stories, got %d", len(rows))
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

func TestSetStoryStatus_Done_RemovesFromBacklog(t *testing.T) {
	root, s := newFixture(t)

	obj := unmarshalObject(t, callTool(t, s, "set_story_status", map[string]any{
		"story_id": "STORY-001",
		"status":   "done",
	}))

	if obj["backlog_removed"] != true {
		t.Errorf("expected backlog_removed=true, got %v", obj["backlog_removed"])
	}

	backlog, _ := os.ReadFile(filepath.Join(root, "backlog.md"))
	if strings.Contains(string(backlog), "STORY-001") {
		t.Error("STORY-001 should have been removed from backlog.md")
	}
	// Remaining entry should be renumbered to 1.
	if !strings.Contains(string(backlog), "1. **STORY-002**") {
		t.Errorf("remaining entry not renumbered:\n%s", backlog)
	}
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
		"status":   "done",
	})
	assertError(t, result, "STORY-999")
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
