package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/corbym/backlog-mcp/parser"
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
			mcp.WithDescription("List stories from the project index, optionally filtered by epic, status, or type. Returns an array of {story_id, title, status, epic_id, story_type} objects. With no filters, returns all stories across all epics."),
			mcp.WithString("epic_id",
				mcp.Description("Optional epic ID to filter by (e.g. EPIC-003). When provided, only stories belonging to this epic are returned."),
			),
			mcp.WithString("status",
				mcp.Description("Optional status to filter by. Valid values: draft, in-progress, done, blocked. When provided, only stories with this status are returned."),
			),
			mcp.WithString("story_type",
				mcp.Description("Optional story type to filter by. Valid values: feature, bug, chore, spike. When provided, only stories of this type are returned."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			epicFilter := strings.ToUpper(optionalString(req, "epic_id"))
			statusFilter := strings.ToLower(optionalString(req, "status"))
			typeFilter := strings.ToLower(optionalString(req, "story_type"))

			epics, err := parser.ParseIndex(cfg.StoriesRoot)
			if err != nil {
				return toolError(err), nil
			}

			type row struct {
				StoryID   string `json:"story_id"`
				Title     string `json:"title"`
				Status    string `json:"status"`
				EpicID    string `json:"epic_id"`
				StoryType string `json:"story_type"`
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
					if typeFilter != "" && s.StoryType != typeFilter {
						continue
					}
					results = append(results, row{
						StoryID:   s.ID,
						Title:     s.Title,
						Status:    s.Status,
						EpicID:    s.EpicID,
						StoryType: s.StoryType,
					})
				}
			}

			return toolJSON(results)
		},
	)

	// ── get_story ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_story",
			mcp.WithDescription("Get the full markdown content and metadata for a single story. Returns {story_id, title, status, epic_id, path, content} where content is the raw markdown of the story file."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to retrieve, e.g. STORY-047"),
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
				result["story_type"] = meta.StoryType
			}

			return toolJSON(result)
		},
	)

	// ── set_story_status ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("set_story_status",
			mcp.WithDescription("Update the status of a story to draft, in-progress, or blocked. "+
				"To mark a story done, use complete_story instead — it enforces acceptance criteria, appends a summary note, and removes the story from the backlog. "+
				"Returns {story_id, old_status, new_status, backlog_updated}."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to update, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithString("status",
				mcp.Description("New status to assign. Must be one of: draft, in-progress, blocked. To mark done, use complete_story."),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			storyID := strings.ToUpper(requiredString(req, "story_id"))
			newStatus := strings.ToLower(requiredString(req, "status"))

			if newStatus == "done" {
				return toolError(fmt.Errorf("use complete_story to mark a story done — it enforces acceptance criteria, appends a summary note, and updates the backlog")), nil
			}
			validStatuses := map[string]bool{
				"draft": true, "in-progress": true, "blocked": true,
			}
			if !validStatuses[newStatus] {
				return toolError(fmt.Errorf("invalid status %q: must be draft, in-progress, or blocked", newStatus)), nil
			}

			// 1. Update requirements-index.md
			oldStatus, err := parser.UpdateStoryStatus(cfg.StoriesRoot, storyID, newStatus)
			if err != nil {
				return toolError(err), nil
			}

			// 2. Update inline status marker in backlog
			backlogUpdated := false
			var backlogWarning string
			if err := parser.UpdateBacklogStatus(cfg.StoriesRoot, storyID, newStatus); err != nil {
				backlogWarning = err.Error()
			} else {
				backlogUpdated = true
			}

			resp := map[string]any{
				"story_id":        storyID,
				"old_status":      oldStatus,
				"new_status":      newStatus,
				"backlog_updated": backlogUpdated,
			}
			if backlogWarning != "" {
				resp["backlog_warning"] = backlogWarning
			}
			return toolJSON(resp)
		},
	)

	// ── add_story_note ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("add_story_note",
			mcp.WithDescription("Append a timestamped note to a story file. Use to record progress, decisions made, or blockers encountered. Notes are appended under a '## Notes' section with an ISO 8601 timestamp. Returns {story_id, appended_at, path}."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to annotate, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithString("note",
				mcp.Description("The note text to append. Can be multi-line. Will be stored with a UTC timestamp."),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

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

	// ── complete_story ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("complete_story",
			mcp.WithDescription("Mark a story done and append a mandatory completion summary note in one atomic call. "+
				"Validates acceptance criteria before completing: if the AC section has not been set (contains only the placeholder), "+
				"completion is blocked — call set_acceptance_criteria first. "+
				"If some criteria are unchecked, incomplete_items is required with one explanation per unchecked item (in the order they appear). "+
				"On success, removes the story from backlog.md and returns {story_id, completed_at, backlog_removed}."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to complete, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithString("summary",
				mcp.Description("Completion summary describing what was done. Appended as a timestamped note to the story file."),
				mcp.Required(),
			),
			mcp.WithArray("incomplete_items",
				mcp.Description("Required when the story has unchecked acceptance criteria. Provide one explanation string per unchecked item, in the order they appear in the story file."),
				mcp.Items(map[string]any{"type": "string"}),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			storyID := strings.ToUpper(requiredString(req, "story_id"))
			summary := strings.TrimSpace(requiredString(req, "summary"))
			if summary == "" {
				return toolError(fmt.Errorf("missing required parameter \"summary\"")), nil
			}

			epics, err := parser.ParseIndex(cfg.StoriesRoot)
			if err != nil {
				return toolError(err), nil
			}

			found := false
			status := ""
			storyEpicID := ""
			for _, epic := range epics {
				for _, s := range epic.Stories {
					if s.ID == storyID {
						found = true
						status = s.Status
						storyEpicID = epic.ID
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				return toolError(fmt.Errorf("story %s not found in index", storyID)), nil
			}
			if status == "done" {
				return toolError(fmt.Errorf("story %s is already done", storyID)), nil
			}

			relPath, err := parser.FindStoryPath(cfg.StoriesRoot, storyID)
			if err != nil {
				return toolError(fmt.Errorf("finding story %s: %w", storyID, err)), nil
			}

			// Validate acceptance criteria before mutating anything.
			acItems, err := parser.ParseAcceptanceCriteria(cfg.StoriesRoot, relPath)
			if err != nil {
				return toolError(err), nil
			}
			isPlaceholder := len(acItems) == 0 ||
				(len(acItems) == 1 && !acItems[0].Checked && acItems[0].Text == "Define acceptance criteria")
			if isPlaceholder {
				return toolError(fmt.Errorf(
					"acceptance criteria have not been set for %s; call set_acceptance_criteria before completing",
					storyID,
				)), nil
			}

			var unchecked []parser.ACItem
			for _, item := range acItems {
				if !item.Checked {
					unchecked = append(unchecked, item)
				}
			}

			var incompleteItems []string
			if len(unchecked) > 0 {
				incompleteItems = optionalStringSlice(req, "incomplete_items")
				if len(incompleteItems) == 0 {
					uncheckedLines := make([]string, len(unchecked))
					for i, u := range unchecked {
						uncheckedLines[i] = "- [ ] " + u.Text
					}
					return toolError(fmt.Errorf(
						"story %s has %d unchecked criterion/criteria; provide incomplete_items with one explanation per unchecked item:\n%s",
						storyID, len(unchecked), strings.Join(uncheckedLines, "\n"),
					)), nil
				}
				if len(incompleteItems) != len(unchecked) {
					return toolError(fmt.Errorf(
						"incomplete_items has %d entries but there are %d unchecked criteria; provide one explanation per unchecked item",
						len(incompleteItems), len(unchecked),
					)), nil
				}
			}

			// Build the note, appending incomplete-criteria details if needed.
			var noteBuilder strings.Builder
			noteBuilder.WriteString(summary)
			if len(incompleteItems) > 0 {
				noteBuilder.WriteString("\n\nIncomplete criteria:\n")
				for i, u := range unchecked {
					noteBuilder.WriteString(fmt.Sprintf("- [ ] %s: %s\n", u.Text, incompleteItems[i]))
				}
			}
			note := noteBuilder.String()

			if _, err := parser.UpdateStoryStatus(cfg.StoriesRoot, storyID, "done"); err != nil {
				return toolError(err), nil
			}

			backlogRemoved := false
			entries, err := parser.ParseBacklog(cfg.StoriesRoot)
			if err != nil {
				return toolError(err), nil
			}
			for _, e := range entries {
				if e.StoryID != storyID {
					continue
				}
				if err := parser.RemoveFromBacklog(cfg.StoriesRoot, storyID); err != nil {
					return toolError(err), nil
				}
				backlogRemoved = true
				break
			}

			if storyEpicID != "" {
				// non-fatal: mark the story done in the epic file
				_ = parser.MarkEpicStoryDone(cfg.StoriesRoot, storyEpicID, storyID)
			}

			completedAt := time.Now().UTC().Format(time.RFC3339)
			if err := parser.AppendNote(cfg.StoriesRoot, relPath, completedAt, note); err != nil {
				return toolError(err), nil
			}

			return toolJSON(map[string]any{
				"story_id":        storyID,
				"completed_at":    completedAt,
				"backlog_removed": backlogRemoved,
			})
		},
	)

	// ── create_epic ──────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("create_epic",
			mcp.WithDescription("Create a new epic. Assigns the next EPIC-NNN ID, creates the epic directory and epic.md file, and registers it in requirements-index.md with status draft. Returns {epic_id, path}."),
			mcp.WithString("title",
				mcp.Description("Title of the epic, e.g. 'User Authentication'"),
				mcp.Required(),
			),
			mcp.WithString("description",
				mcp.Description("Optional description or goal for the epic. Written into the epic.md file."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			title := requiredString(req, "title")
			description := optionalString(req, "description")

			epicID, epicDir, err := parser.CreateEpic(cfg.StoriesRoot, title, description)
			if err != nil {
				return toolError(err), nil
			}

			return toolJSON(map[string]any{
				"epic_id": epicID,
				"path":    epicDir,
			})
		},
	)

	// ── create_story ─────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("create_story",
			mcp.WithDescription("Create a new story under an existing epic. Assigns the next STORY-NNN ID, writes the story file, and registers it in requirements-index.md and backlog.md with status draft. The story is appended to the end of the backlog. Returns {story_id, path}."),
			mcp.WithString("epic_id",
				mcp.Description("Epic ID the story belongs to, e.g. EPIC-003. The epic must already exist."),
				mcp.Required(),
			),
			mcp.WithString("title",
				mcp.Description("Title of the story, e.g. 'User can reset password'"),
				mcp.Required(),
			),
			mcp.WithString("description",
				mcp.Description("Optional description or goal for the story. Written into the story.md file."),
			),
			mcp.WithString("story_type",
				mcp.Description("Type of story. Valid values: feature, bug, chore, spike. Defaults to 'feature' if not provided."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			epicID := strings.ToUpper(requiredString(req, "epic_id"))
			title := requiredString(req, "title")
			description := optionalString(req, "description")
			storyType := strings.ToLower(optionalString(req, "story_type"))
			if storyType == "" {
				storyType = "feature"
			}
			validTypes := map[string]bool{"feature": true, "bug": true, "chore": true, "spike": true}
			if !validTypes[storyType] {
				return toolError(fmt.Errorf("invalid story_type %q: must be feature, bug, chore, or spike", storyType)), nil
			}

			storyID, relPath, err := parser.CreateStory(cfg.StoriesRoot, epicID, title, description, storyType)
			if err != nil {
				return toolError(err), nil
			}

			return toolJSON(map[string]any{
				"story_id":   storyID,
				"path":       relPath,
				"story_type": storyType,
			})
		},
	)

	// ── set_acceptance_criteria ──────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("set_acceptance_criteria",
			mcp.WithDescription("Replace the acceptance criteria section of a story file. Each string in the criteria array becomes a `- [ ] ...` checklist line. Idempotent: calling again replaces the previous AC entirely. Acceptance criteria must be set before a story can be completed with complete_story. Returns {story_id, criteria_count, path}."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to update, e.g. STORY-007"),
				mcp.Required(),
			),
			mcp.WithArray("criteria",
				mcp.Description("List of acceptance criteria strings. Each entry becomes a checklist item (- [ ] ...) in the story file. Must contain at least one item."),
				mcp.Items(map[string]any{"type": "string"}),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			storyID := strings.ToUpper(requiredString(req, "story_id"))

			criteria, err := requiredStringSlice(req, "criteria")
			if err != nil {
				return toolError(err), nil
			}
			if len(criteria) == 0 {
				return toolError(fmt.Errorf("criteria must not be empty")), nil
			}

			relPath, err := parser.FindStoryPath(cfg.StoriesRoot, storyID)
			if err != nil {
				return toolError(fmt.Errorf("finding story %s: %w", storyID, err)), nil
			}

			if err := parser.SetAcceptanceCriteria(cfg.StoriesRoot, relPath, criteria); err != nil {
				return toolError(err), nil
			}

			return toolJSON(map[string]any{
				"story_id":       storyID,
				"criteria_count": len(criteria),
				"path":           relPath,
			})
		},
	)

	// ── check_acceptance_criterion ───────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("check_acceptance_criterion",
			mcp.WithDescription("Mark a single acceptance criterion as checked (- [ ] → - [x]) in a story file. "+
				"Identify the target by criterion_index (0-based) or criterion_text (case-insensitive exact match). "+
				"Exactly one must be provided. "+
				"Returns {story_id, criterion, checked, path}. "+
				"Errors if the story is not found, the criterion is not found, or it is already checked."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to update, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithNumber("criterion_index",
				mcp.Description("0-based index of the criterion to check. Use when you know the position. Mutually exclusive with criterion_text."),
			),
			mcp.WithString("criterion_text",
				mcp.Description("Exact text of the criterion to check (case-insensitive). Use when you know the text. Mutually exclusive with criterion_index."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			storyID := strings.ToUpper(requiredString(req, "story_id"))
			criterionIndex := req.GetInt("criterion_index", -1)
			criterionText := optionalString(req, "criterion_text")

			if criterionIndex < 0 && criterionText == "" {
				return toolError(fmt.Errorf("must provide either criterion_index or criterion_text")), nil
			}

			relPath, err := parser.FindStoryPath(cfg.StoriesRoot, storyID)
			if err != nil {
				return toolError(fmt.Errorf("story %s not found: %w", storyID, err)), nil
			}

			checkedText, err := parser.CheckAcceptanceCriterion(cfg.StoriesRoot, relPath, criterionIndex, criterionText)
			if err != nil {
				return toolError(err), nil
			}

			return toolJSON(map[string]any{
				"story_id":  storyID,
				"criterion": checkedText,
				"checked":   true,
				"path":      relPath,
			})
		},
	)

	// ── groom_epic ───────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("groom_epic",
			mcp.WithDescription("Reconcile the ## Stories section in an epic.md file with the story files on disk and the requirements index. "+
				"Adds missing entries, removes entries for story files that no longer exist, and refreshes titles and done/undone markers. "+
				"Returns {epic_id, added, removed, updated, unchanged}."),
			mcp.WithString("epic_id",
				mcp.Description("Epic ID to groom, e.g. EPIC-003"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			epicID := strings.ToUpper(requiredString(req, "epic_id"))

			result, err := parser.GroomEpic(cfg.StoriesRoot, epicID)
			if err != nil {
				return toolError(err), nil
			}

			return toolJSON(result)
		},
	)

	// ── get_index_summary ────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_index_summary",
			mcp.WithDescription("Get a high-level summary of all epics and their story counts broken down by status. Useful for situational awareness at the start of a session, without reading every file. Returns an array of {epic_id, title, status, counts: {status: n}, stories: [{story_id, status}]}."),
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
	return req.GetString(key, "")
}

func optionalString(req mcp.CallToolRequest, key string) string {
	return req.GetString(key, "")
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

// optionalStringSlice extracts an optional array-of-strings parameter from a tool request.
// Returns nil if the parameter is absent or not an array of strings.
func optionalStringSlice(req mcp.CallToolRequest, key string) []string {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil
	}
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil
		}
		result = append(result, s)
	}
	return result
}

// requiredStringSlice extracts a required array-of-strings parameter from a tool request.
func requiredStringSlice(req mcp.CallToolRequest, key string) ([]string, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing required parameter %q", key)
	}
	v, ok := args[key]
	if !ok || v == nil {
		return nil, fmt.Errorf("missing required parameter %q", key)
	}
	raw, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("parameter %q must be an array of strings", key)
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("parameter %q must be an array of strings, got non-string element", key)
		}
		result = append(result, s)
	}
	return result, nil
}
