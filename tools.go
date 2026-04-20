package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/corbym/backlog-mcp/parser"

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
			mcp.WithDescription("Update the status of a story to draft, in-progress, blocked, or deferred. "+
				"To mark a story done, use complete_story instead — it enforces acceptance criteria, appends a summary note, and removes the story from the backlog. "+
				"Returns {story_id, old_status, new_status, backlog_updated}."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to update, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithString("status",
				mcp.Description("New status to assign. Must be one of: draft, in-progress, blocked, deferred. To mark done, use complete_story."),
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
				"draft": true, "in-progress": true, "blocked": true, "deferred": true,
			}
			if !validStatuses[newStatus] {
				return toolError(fmt.Errorf("invalid status %q: must be draft, in-progress, blocked, or deferred", newStatus)), nil
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

			// 3. If the story markdown has a status metadata line, keep it in sync.
			storyMetadataUpdated := false
			var storyMetadataWarning string
			if relPath, err := parser.FindStoryPath(cfg.StoriesRoot, storyID); err != nil {
				storyMetadataWarning = err.Error()
			} else if updated, err := parser.UpdateStoryStatusMetadata(cfg.StoriesRoot, relPath, newStatus); err != nil {
				storyMetadataWarning = err.Error()
			} else {
				storyMetadataUpdated = updated
			}

			resp := map[string]any{
				"story_id":               storyID,
				"old_status":             oldStatus,
				"new_status":             newStatus,
				"backlog_updated":        backlogUpdated,
				"story_metadata_updated": storyMetadataUpdated,
			}
			if backlogWarning != "" {
				resp["backlog_warning"] = backlogWarning
			}
			if storyMetadataWarning != "" {
				resp["story_metadata_warning"] = storyMetadataWarning
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
				"IMPORTANT: if a criterion is actually done, mark it [x] in the story file via set_acceptance_criteria BEFORE calling this tool — do not leave it unchecked. "+
				"If criteria remain unchecked (genuinely not done), incomplete_items is required with one explanation per unchecked item explaining WHY it was not completed (e.g. deferred, out of scope). "+
				"incomplete_items is for unfinished work only — never use it to confirm completed work. "+
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
				mcp.Description("Required when the story has unchecked (genuinely unfinished) acceptance criteria. "+
					"Each string must explain WHY that criterion was not met (e.g. 'Deferred to STORY-010 — rarity system not yet designed'). "+
					"One entry per unchecked item, in the order they appear. "+
					"DO NOT use this field to confirm items that are done — if a criterion is done, tick it [x] via set_acceptance_criteria first, then retry. "+
					"Never prefix entries with 'Done:' — if it is done, it should not appear here at all."),
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
						"story %s has %d unchecked criterion/criteria:\n%s\n\nIf these are actually done, mark them [x] via set_acceptance_criteria first, then retry complete_story with no incomplete_items.\nIf they are genuinely not done, provide incomplete_items with one explanation per item explaining WHY it was not completed (not a confirmation that it was).",
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

	// ── set_epic_status ──────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("set_epic_status",
			mcp.WithDescription("Update the lifecycle status of an epic. "+
				"Use this tool to manage the epic's own status — not the status of individual stories within it (use set_story_status for that). "+
				"Typical progression: draft → in-progress (when the first story starts) → done (when all stories are complete) or deferred (if the epic is postponed). "+
				"Status meanings: "+
				"'draft' = epic created but no work started; "+
				"'in-progress' = actively being worked on; "+
				"'done' = all stories complete and the epic is closed; "+
				"'blocked' = progress prevented by an external dependency; "+
				"'deferred' = postponed indefinitely. "+
				"Guards: "+
				"(1) Setting 'done' requires a summary and checks all stories are done. If any are not done, the call fails — set override_incomplete=true only after the user explicitly confirms this is acceptable. "+
				"(2) Moving backwards (e.g. done → in-progress, in-progress → draft) asks you to create new stories to justify the regression first. Set confirm_regression=true only if the user explicitly insists on skipping story creation. "+
				"Returns {epic_id, old_status, new_status}."),
			mcp.WithString("epic_id",
				mcp.Description("Epic ID to update, e.g. EPIC-003"),
				mcp.Required(),
			),
			mcp.WithString("status",
				mcp.Description("New status to assign. Must be one of: draft, in-progress, done, blocked, deferred."),
				mcp.Required(),
			),
			mcp.WithString("summary",
				mcp.Description("Required when setting status to 'done'. Describes what was accomplished by this epic. Appended as a timestamped note to the epic file."),
			),
			mcp.WithBoolean("override_incomplete",
				mcp.Description("Set to true to mark the epic 'done' even when some stories are not done. Only set after the user explicitly confirms the incomplete stories are intentionally omitted."),
			),
			mcp.WithBoolean("confirm_regression",
				mcp.Description("Set to true to allow a backwards status transition (e.g. done → in-progress) without first creating new stories. Only set if the user explicitly insists on skipping story creation."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			epicID := strings.ToUpper(requiredString(req, "epic_id"))
			newStatus := strings.ToLower(requiredString(req, "status"))
			summary := strings.TrimSpace(optionalString(req, "summary"))
			overrideIncomplete := req.GetBool("override_incomplete", false)
			confirmRegression := req.GetBool("confirm_regression", false)

			validStatuses := map[string]bool{
				"draft": true, "in-progress": true, "done": true, "blocked": true, "deferred": true,
			}
			if !validStatuses[newStatus] {
				return toolError(fmt.Errorf("invalid status %q: must be draft, in-progress, done, blocked, or deferred", newStatus)), nil
			}

			epics, err := parser.ParseIndex(cfg.StoriesRoot)
			if err != nil {
				return toolError(err), nil
			}

			var targetEpic *parser.Epic
			for i := range epics {
				if epics[i].ID == epicID {
					targetEpic = &epics[i]
					break
				}
			}
			if targetEpic == nil {
				return toolError(fmt.Errorf("epic %s not found in index", epicID)), nil
			}

			oldStatus := targetEpic.Status

			// Guard: require summary and check story completion when marking done.
			if newStatus == "done" {
				if summary == "" {
					return toolError(fmt.Errorf("summary is required when setting status to 'done'")), nil
				}
				if !overrideIncomplete {
					var notDone []string
					for _, s := range targetEpic.Stories {
						if s.Status != "done" {
							notDone = append(notDone, fmt.Sprintf("  - %s (%s): %s", s.ID, s.Status, s.Title))
						}
					}
					if len(notDone) > 0 {
						return toolError(fmt.Errorf(
							"%s has %d story/stories not yet done:\n%s\n\nComplete them first, or set override_incomplete=true if the user has confirmed these are intentionally omitted.",
							epicID, len(notDone), strings.Join(notDone, "\n"),
						)), nil
					}
				}
			}

			// Guard: prompt to add stories before allowing a backwards transition.
			statusRank := map[string]int{"draft": 0, "in-progress": 1, "done": 2}
			oldRank, oldRanked := statusRank[oldStatus]
			newRank, newRanked := statusRank[newStatus]
			if oldRanked && newRanked && oldRank > newRank && !confirmRegression {
				return toolError(fmt.Errorf(
					"%s is moving backwards from '%s' to '%s'. "+
						"Please create new stories for the remaining work first, then call set_epic_status again. "+
						"Set confirm_regression=true to proceed without adding stories if the user explicitly insists.",
					epicID, oldStatus, newStatus,
				)), nil
			}

			if _, err := parser.UpdateEpicStatus(cfg.StoriesRoot, epicID, newStatus); err != nil {
				return toolError(err), nil
			}

			if summary != "" {
				epicRelPath, err := parser.FindEpicFilePath(cfg.StoriesRoot, epicID)
				if err == nil {
					ts := time.Now().UTC().Format(time.RFC3339)
					_ = parser.AppendNote(cfg.StoriesRoot, epicRelPath, ts, summary)
				}
			}

			return toolJSON(map[string]any{
				"epic_id":    epicID,
				"old_status": oldStatus,
				"new_status": newStatus,
			})
		},
	)

	// ── bulk_update_acceptance_criteria ─────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("bulk_update_acceptance_criteria",
			mcp.WithDescription("Update the checked state of individual acceptance criteria on a story in one operation. "+
				"Only the criteria explicitly listed are modified; all others are left untouched. "+
				"Criteria are matched by exact text. If any criterion text is not found, no changes are made and an error is returned. "+
				"Returns {story_id, path, content, criteria_updated, errors}."),
			mcp.WithString("story_id",
				mcp.Description("Story ID to update, e.g. STORY-047"),
				mcp.Required(),
			),
			mcp.WithObject("criteria",
				mcp.Description("Map of criterion text to desired checked state. true = checked [x], false = unchecked [ ]. Criterion text must match exactly."),
				mcp.AdditionalProperties(map[string]any{"type": "boolean"}),
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
			if storyID == "" {
				return toolError(fmt.Errorf("missing required parameter \"story_id\"")), nil
			}

			criteriaMap, err := requiredBoolMap(req, "criteria")
			if err != nil {
				return toolError(err), nil
			}
			if len(criteriaMap) == 0 {
				return toolError(fmt.Errorf("criteria must not be empty")), nil
			}

			relPath, err := parser.FindStoryPath(cfg.StoriesRoot, storyID)
			if err != nil {
				return toolError(fmt.Errorf("story %s not found: %w", storyID, err)), nil
			}

			notFound, err := parser.PatchAcceptanceCriteria(cfg.StoriesRoot, relPath, criteriaMap)
			if err != nil {
				return toolError(err), nil
			}

			content, err := parser.ReadStory(cfg.StoriesRoot, relPath)
			if err != nil {
				return toolError(err), nil
			}

			updated := make([]string, 0, len(criteriaMap))
			for text := range criteriaMap {
				updated = append(updated, text)
			}

			return toolJSON(map[string]any{
				"story_id":         storyID,
				"path":             relPath,
				"content":          content,
				"criteria_updated": updated,
				"errors":           notFound,
			})
		},
	)

	// ── bulk_update_stories ──────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("bulk_update_stories",
			mcp.WithDescription("Update multiple stories in one operation. Each entry may set status, append a note, and/or patch acceptance criteria. "+
				"Updates are applied atomically per file. If a story does not exist, an error is recorded for that entry and processing continues. "+
				"Returns an array of per-story result objects with fields: story_id, status_updated, old_status, new_status, note_appended, criteria_updated, criteria_errors, errors."),
			mcp.WithArray("updates",
				mcp.Description("Array of story update objects. Each must include story_id; status, note, and criteria are optional. "+
					"status must be one of: draft, in-progress, blocked, deferred (use complete_story to mark done). "+
					"note is appended, not replaced. "+
					"criteria is a map of criterion text to boolean checked state."),
				mcp.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"story_id": map[string]any{"type": "string"},
						"status":   map[string]any{"type": "string"},
						"note":     map[string]any{"type": "string"},
						"criteria": map[string]any{
							"type":                 "object",
							"additionalProperties": map[string]any{"type": "boolean"},
						},
					},
					"required": []any{"story_id"},
				}),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			updates, err := requiredObjectSlice(req, "updates")
			if err != nil {
				return toolError(err), nil
			}

			type rowResult struct {
				StoryID         string   `json:"story_id"`
				StatusUpdated   bool     `json:"status_updated"`
				OldStatus       string   `json:"old_status,omitempty"`
				NewStatus       string   `json:"new_status,omitempty"`
				NoteAppended    bool     `json:"note_appended"`
				CriteriaUpdated []string `json:"criteria_updated,omitempty"`
				CriteriaErrors  []string `json:"criteria_errors,omitempty"`
				Errors          []string `json:"errors"`
			}

			results := make([]rowResult, 0, len(updates))
			validStatuses := map[string]bool{
				"draft": true, "in-progress": true, "blocked": true, "deferred": true,
			}

			for _, upd := range updates {
				storyID := strings.ToUpper(stringField(upd, "story_id"))
				row := rowResult{StoryID: storyID, Errors: []string{}}

				if storyID == "" {
					row.Errors = append(row.Errors, "story_id is required")
					results = append(results, row)
					continue
				}

				relPath, pathErr := parser.FindStoryPath(cfg.StoriesRoot, storyID)
				if pathErr != nil {
					row.Errors = append(row.Errors, fmt.Sprintf("story %s not found", storyID))
					results = append(results, row)
					continue
				}

				// Optional: status update
				if rawStatus, ok := upd["status"]; ok && rawStatus != nil {
					newStatus := strings.ToLower(fmt.Sprintf("%v", rawStatus))
					if newStatus == "done" {
						row.Errors = append(row.Errors, "use complete_story to mark a story done")
					} else if !validStatuses[newStatus] {
						row.Errors = append(row.Errors, fmt.Sprintf("invalid status %q: must be draft, in-progress, blocked, or deferred", newStatus))
					} else {
						oldStatus, statusErr := parser.UpdateStoryStatus(cfg.StoriesRoot, storyID, newStatus)
						if statusErr != nil {
							row.Errors = append(row.Errors, statusErr.Error())
						} else {
							row.StatusUpdated = true
							row.OldStatus = oldStatus
							row.NewStatus = newStatus
							_ = parser.UpdateBacklogStatus(cfg.StoriesRoot, storyID, newStatus)
							if _, err := parser.UpdateStoryStatusMetadata(cfg.StoriesRoot, relPath, newStatus); err != nil {
								// non-fatal
							}
						}
					}
				}

				// Optional: append note
				if rawNote, ok := upd["note"]; ok && rawNote != nil {
					note := strings.TrimSpace(fmt.Sprintf("%v", rawNote))
					if note != "" {
						ts := time.Now().UTC().Format(time.RFC3339)
						if noteErr := parser.AppendNote(cfg.StoriesRoot, relPath, ts, note); noteErr != nil {
							row.Errors = append(row.Errors, noteErr.Error())
						} else {
							row.NoteAppended = true
						}
					}
				}

				// Optional: patch criteria
				if rawCriteria, ok := upd["criteria"]; ok && rawCriteria != nil {
					criteriaMap, mapErr := extractBoolMap(rawCriteria)
					if mapErr != nil {
						row.Errors = append(row.Errors, mapErr.Error())
					} else if len(criteriaMap) > 0 {
						notFound, patchErr := parser.PatchAcceptanceCriteria(cfg.StoriesRoot, relPath, criteriaMap)
						if patchErr != nil {
							row.CriteriaErrors = notFound
						} else {
							updated := make([]string, 0, len(criteriaMap))
							for text := range criteriaMap {
								updated = append(updated, text)
							}
							row.CriteriaUpdated = updated
						}
					}
				}

				results = append(results, row)
			}

			return toolJSON(results)
		},
	)

	// ── bulk_update_epics ────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("bulk_update_epics",
			mcp.WithDescription("Update multiple epics in one operation. Each entry may set status and/or append a note. "+
				"Updates are applied atomically per file. If an epic does not exist, an error is recorded for that entry and processing continues. "+
				"Returns an array of per-epic result objects with fields: epic_id, status_updated, old_status, new_status, note_appended, errors."),
			mcp.WithArray("updates",
				mcp.Description("Array of epic update objects. Each must include epic_id; status and note are optional. "+
					"status must be one of: draft, in-progress, done, blocked, deferred. "+
					"note is appended, not replaced."),
				mcp.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"epic_id": map[string]any{"type": "string"},
						"status":  map[string]any{"type": "string"},
						"note":    map[string]any{"type": "string"},
					},
					"required": []any{"epic_id"},
				}),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			unlock, err := parser.AcquireLock(cfg.StoriesRoot, 5*time.Second)
			if err != nil {
				return toolError(err), nil
			}
			defer unlock()

			updates, err := requiredObjectSlice(req, "updates")
			if err != nil {
				return toolError(err), nil
			}

			type rowResult struct {
				EpicID        string   `json:"epic_id"`
				StatusUpdated bool     `json:"status_updated"`
				OldStatus     string   `json:"old_status,omitempty"`
				NewStatus     string   `json:"new_status,omitempty"`
				NoteAppended  bool     `json:"note_appended"`
				Errors        []string `json:"errors"`
			}

			validStatuses := map[string]bool{
				"draft": true, "in-progress": true, "done": true, "blocked": true, "deferred": true,
			}

			results := make([]rowResult, 0, len(updates))

			for _, upd := range updates {
				epicID := strings.ToUpper(stringField(upd, "epic_id"))
				row := rowResult{EpicID: epicID, Errors: []string{}}

				if epicID == "" {
					row.Errors = append(row.Errors, "epic_id is required")
					results = append(results, row)
					continue
				}

				// Optional: status update
				if rawStatus, ok := upd["status"]; ok && rawStatus != nil {
					newStatus := strings.ToLower(fmt.Sprintf("%v", rawStatus))
					if !validStatuses[newStatus] {
						row.Errors = append(row.Errors, fmt.Sprintf("invalid status %q: must be draft, in-progress, done, blocked, or deferred", newStatus))
					} else {
						oldStatus, statusErr := parser.UpdateEpicStatus(cfg.StoriesRoot, epicID, newStatus)
						if statusErr != nil {
							row.Errors = append(row.Errors, statusErr.Error())
						} else {
							row.StatusUpdated = true
							row.OldStatus = oldStatus
							row.NewStatus = newStatus
						}
					}
				}

				// Optional: append note
				if rawNote, ok := upd["note"]; ok && rawNote != nil {
					note := strings.TrimSpace(fmt.Sprintf("%v", rawNote))
					if note != "" {
						epicRelPath, pathErr := parser.FindEpicFilePath(cfg.StoriesRoot, epicID)
						if pathErr != nil {
							row.Errors = append(row.Errors, pathErr.Error())
						} else {
							ts := time.Now().UTC().Format(time.RFC3339)
							if noteErr := parser.AppendNote(cfg.StoriesRoot, epicRelPath, ts, note); noteErr != nil {
								row.Errors = append(row.Errors, noteErr.Error())
							} else {
								row.NoteAppended = true
							}
						}
					}
				}

				// If neither status nor note was provided and no errors either, the epic_id
				// must exist — validate that now (the status update would have caught it already).
				if !row.StatusUpdated && !row.NoteAppended && len(row.Errors) == 0 {
					if _, pathErr := parser.FindEpicFilePath(cfg.StoriesRoot, epicID); pathErr != nil {
						row.Errors = append(row.Errors, fmt.Sprintf("epic %s not found", epicID))
					}
				}

				results = append(results, row)
			}

			return toolJSON(results)
		},
	)


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

// requiredBoolMap extracts a required object-of-booleans parameter from a tool request.
func requiredBoolMap(req mcp.CallToolRequest, key string) (map[string]bool, error) {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing required parameter %q", key)
	}
	v, ok := args[key]
	if !ok || v == nil {
		return nil, fmt.Errorf("missing required parameter %q", key)
	}
	return extractBoolMap(v)
}

// extractBoolMap converts a map[string]any (with boolean values) to map[string]bool.
func extractBoolMap(v any) (map[string]bool, error) {
	raw, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected an object with boolean values")
	}
	result := make(map[string]bool, len(raw))
	for k, val := range raw {
		b, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("value for %q must be a boolean, got %T", k, val)
		}
		result[k] = b
	}
	return result, nil
}

// requiredObjectSlice extracts a required array-of-objects parameter from a tool request.
func requiredObjectSlice(req mcp.CallToolRequest, key string) ([]map[string]any, error) {
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
		return nil, fmt.Errorf("parameter %q must be an array of objects", key)
	}
	result := make([]map[string]any, 0, len(raw))
	for i, item := range raw {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("parameter %q item %d must be an object", key, i)
		}
		result = append(result, obj)
	}
	return result, nil
}

// stringField extracts a string field from a map[string]any, returning "" if absent or not a string.
func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

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
