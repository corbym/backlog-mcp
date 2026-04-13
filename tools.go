package backlog

import (
	"backlog/parser"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerTools wires all tool handlers onto the MCP server.
func registerTools(s *server.MCPServer, cfg *Config) {
	// ── list_stories ────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("list_stories",
			mcp.WithDescription("List stories from the project index, optionally filtered by epic or status."),
			mcp.WithString("epic_id",
				mcp.Description("Optional epic ID to filter by, e.g. EPIC-003"),
			),
			mcp.WithString("status",
				mcp.Description("Optional status to filter by, e.g. draft, in-progress, done"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			epicFilter := strings.ToUpper(optionalString(req, "epic_id"))
			statusFilter := strings.ToLower(optionalString(req, "status"))

			epics, err := parser.ParseIndex(cfg.StoriesRoot)
			if err != nil {
				return toolError(err), nil
			}

			type row struct {
				StoryID string `json:"story_id"`
				Title   string `json:"title"`
				Status  string `json:"status"`
				EpicID  string `json:"epic_id"`
			}
			var results []row

			for _, epic := range epics {
				if epicFilter != "" && epic.ID != epicFilter {
					continue
				}
				for _, s := range epic.Stories {
					if statusFilter != "" && s.Status != statusFilter {
						continue
					}
					results = append(results, row{
						StoryID: s.ID,
						Title:   s.Title,
						Status:  s.Status,
						EpicID:  s.EpicID,
					})
				}
			}

			return toolJSON(results)
		},
	)

	// ── get_story ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_story",
			mcp.WithDescription("Get the full markdown content and metadata for a single story."),
			mcp.WithString("story_id",
				mcp.Description("Story ID, e.g. STORY-047"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			storyID := strings.ToUpper(requiredString(req, "story_id"))

			// resolve metadata from index
			epics, err := parser.ParseIndex(cfg.StoriesRoot)
			if err != nil {
				return toolError(err), nil
			}
			var meta *parser.Story
			for _, epic := range epics {
				for _, s := range epic.Stories {
					if s.ID == storyID {
						s := s
						meta = &s
						break
					}
				}
				if meta != nil {
					break
				}
			}

			// resolve file path from filesystem
			relPath, err := parser.FindStoryPath(cfg.StoriesRoot, storyID)
			if err != nil {
				return toolError(fmt.Errorf("finding story file: %w", err)), nil
			}

			content, err := parser.ReadStory(cfg.StoriesRoot, relPath)
			if err != nil {
				return toolError(err), nil
			}

			result := map[string]any{
				"story_id": storyID,
				"path":     relPath,
				"content":  content,
			}
			if meta != nil {
				result["title"] = meta.Title
				result["status"] = meta.Status
				result["epic_id"] = meta.EpicID
			}

			return toolJSON(result)
		},
	)

	// ── set_story_status ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("set_story_status",
			mcp.WithDescription("Update the status of a story in requirements-index.md and backlog.md."),
			mcp.WithString("story_id",
				mcp.Description("Story ID, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithString("status",
				mcp.Description("New status: draft, in-progress, done, or blocked"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			storyID := strings.ToUpper(requiredString(req, "story_id"))
			newStatus := strings.ToLower(requiredString(req, "status"))

			validStatuses := map[string]bool{
				"draft": true, "in-progress": true, "done": true, "blocked": true,
			}
			if !validStatuses[newStatus] {
				return toolError(fmt.Errorf("invalid status %q: must be draft, in-progress, done, or blocked", newStatus)), nil
			}

			// 1. Update requirements-index.md
			oldStatus, err := parser.UpdateStoryStatus(cfg.StoriesRoot, storyID, newStatus)
			if err != nil {
				return toolError(err), nil
			}

			backlogRemoved := false
			backlogUpdated := false

			// 2. Backlog handling
			if newStatus == "done" {
				// Remove from backlog entirely (per backlog.md rules)
				if err := parser.RemoveFromBacklog(cfg.StoriesRoot, storyID); err != nil {
					// Not fatal — story may not be in backlog
					_ = err
				} else {
					backlogRemoved = true
				}
			} else {
				// Update inline status marker in backlog
				if err := parser.UpdateBacklogStatus(cfg.StoriesRoot, storyID, newStatus); err == nil {
					backlogUpdated = true
				}
			}

			return toolJSON(map[string]any{
				"story_id":        storyID,
				"old_status":      oldStatus,
				"new_status":      newStatus,
				"backlog_removed": backlogRemoved,
				"backlog_updated": backlogUpdated,
			})
		},
	)

	// ── add_story_note ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("add_story_note",
			mcp.WithDescription("Append a timestamped note to a story file. Use to record what was done, decisions made, or blockers."),
			mcp.WithString("story_id",
				mcp.Description("Story ID, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithString("note",
				mcp.Description("The note text to append."),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			storyID := strings.ToUpper(requiredString(req, "story_id"))
			note := requiredString(req, "note")

			relPath, err := parser.FindStoryPath(cfg.StoriesRoot, storyID)
			if err != nil {
				return toolError(err), nil
			}

			timestamp := time.Now().UTC().Format(time.RFC3339)
			if err := parser.AppendNote(cfg.StoriesRoot, relPath, timestamp, note); err != nil {
				return toolError(err), nil
			}

			return toolJSON(map[string]any{
				"story_id":    storyID,
				"appended_at": timestamp,
				"path":        relPath,
			})
		},
	)

	// ── get_index_summary ────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_index_summary",
			mcp.WithDescription("Get a high-level summary of all epics and their story counts by status. Useful for situational awareness without reading every file."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			epics, err := parser.ParseIndex(cfg.StoriesRoot)
			if err != nil {
				return toolError(err), nil
			}

			type epicSummary struct {
				EpicID  string              `json:"epic_id"`
				Title   string              `json:"title"`
				Status  string              `json:"status"`
				Counts  map[string]int      `json:"counts"`
				Stories []map[string]string `json:"stories"`
			}

			var result []epicSummary
			for _, epic := range epics {
				counts := map[string]int{}
				var stories []map[string]string
				for _, s := range epic.Stories {
					counts[s.Status]++
					stories = append(stories, map[string]string{
						"story_id": s.ID,
						"status":   s.Status,
					})
				}
				result = append(result, epicSummary{
					EpicID:  epic.ID,
					Title:   epic.Title,
					Status:  epic.Status,
					Counts:  counts,
					Stories: stories,
				})
			}

			return toolJSON(result)
		},
	)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func requiredString(req mcp.CallToolRequest, key string) string {
	v, _ := req.Params.Arguments[key].(string)
	return v
}

func optionalString(req mcp.CallToolRequest, key string) string {
	v, _ := req.Params.Arguments[key].(string)
	return v
}

func toolJSON(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return toolError(err), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func toolError(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}
