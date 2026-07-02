# EPIC-005: Backlog MCP Quality & DX

## Goal

Fixes, documentation improvements, and quality-of-life features for the backlog-mcp server. Covers bugs surfaced during dogfooding, tool discovery gaps in deferred-tool environments (VS Code Copilot), and missing MCP tools.

## Stories

- [ ] [STORY-024](story-024.md) — Add reorder_backlog MCP tool
- [ ] [STORY-025](story-025.md) — Fix em-dash encoding artefacts in acceptance criteria text
- [ ] [STORY-026](story-026.md) — Add tool-surface discovery hint to at least one tool description
- [ ] [STORY-027](story-027.md) — Update AGENTS.md with complete tool surface
- [ ] [STORY-028](story-028.md) — Document bulk_update_stories criteria key format
- [ ] [STORY-029](story-029.md) — Fix backlog.go regex to handle format deviations gracefully
- [ ] [STORY-030](story-030.md) — Make UpdateBacklogStatus loud when story is not in backlog
- [x] [STORY-031](story-031.md) — Fix set_acceptance_criteria corruption when passed pre-ticked checkbox strings
- [ ] [STORY-032](story-032.md) — Fix list_stories Unicode encoding artefacts in story titles
- [ ] [STORY-033](story-033.md) — Fix bulk_update_acceptance_criteria combined ID:text key lookup failure
