# STORY-022: Add bulk update tools for stories, epics, and acceptance criteria

**Type:** feature

## Goal

Add three new MCP tools to reduce token overhead: bulk_update_acceptance_criteria, bulk_update_stories, and bulk_update_epics. Each accepts an array of items and applies updates atomically, collecting per-row errors without aborting the batch.

## Acceptance criteria

- [x] AC-STORY-022-fd1a182d: bulk_update_acceptance_criteria tool accepts story_id and criteria map and patches checked states by exact text match
- [x] AC-STORY-022-5c422385: bulk_update_acceptance_criteria aborts without writing if any criterion text is not found
- [x] AC-STORY-022-f0bced1c: bulk_update_stories tool accepts an array of updates with story_id, status, note, and/or criteria fields
- [x] AC-STORY-022-ec6a40a4: bulk_update_stories blocks status=done and directs the caller to use complete_story
- [x] AC-STORY-022-48cdbbe5: bulk_update_stories collects per-row errors without aborting the whole batch
- [x] AC-STORY-022-6c46d6d2: bulk_update_epics tool accepts an array of updates with epic_id, status, and/or note fields
- [x] AC-STORY-022-f598894f: bulk_update_epics collects per-row errors without aborting the whole batch
- [x] AC-STORY-022-7e4fcb97: All three tools are registered in server.go and documented in AGENTS.md
- [x] AC-STORY-022-5300e12e: All writes use the atomic write pattern (temp file + rename)
- [x] AC-STORY-022-db4a3e9c: Tests exist for all three new tools

## Notes

<!-- backlog-mcp: 2026-04-20T22:14:26Z -->
PR #7: Copilot/story 022 add bulk update tools

<!-- backlog-mcp: 2026-05-03T21:41:01Z -->
All three bulk tools implemented, tested, and documented: bulk_update_acceptance_criteria (patches checked states by exact text, aborts on missing criterion), bulk_update_stories (updates status/note/criteria per row, blocks done, collects per-row errors), bulk_update_epics (updates status/note per epic, collects per-row errors). All registered in server.go, documented in AGENTS.md, covered by tests, using atomic writes throughout. All tests passing.
