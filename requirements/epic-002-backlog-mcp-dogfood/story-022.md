# STORY-022: Add bulk update tools for stories, epics, and acceptance criteria

**Type:** feature

## Goal

Add three new MCP tools to reduce token overhead: bulk_update_acceptance_criteria, bulk_update_stories, and bulk_update_epics. Each accepts an array of items and applies updates atomically, collecting per-row errors without aborting the batch.

## Acceptance criteria

- [ ] bulk_update_acceptance_criteria tool accepts story_id and criteria map and patches checked states by exact text match
- [ ] bulk_update_acceptance_criteria aborts without writing if any criterion text is not found
- [ ] bulk_update_stories tool accepts an array of updates with story_id, status, note, and/or criteria fields
- [ ] bulk_update_stories blocks status=done and directs the caller to use complete_story
- [ ] bulk_update_stories collects per-row errors without aborting the whole batch
- [ ] bulk_update_epics tool accepts an array of updates with epic_id, status, and/or note fields
- [ ] bulk_update_epics collects per-row errors without aborting the whole batch
- [ ] All three tools are registered in server.go and documented in AGENTS.md
- [ ] All writes use the atomic write pattern (temp file + rename)
- [ ] Tests exist for all three new tools

## Notes

<!-- backlog-mcp: 2026-04-20T22:14:26Z -->
PR #7: Copilot/story 022 add bulk update tools
