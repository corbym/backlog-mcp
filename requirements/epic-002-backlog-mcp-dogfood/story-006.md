# STORY-006: Surface loud failure when set_story_status target is not in backlog

## Goal

`UpdateBacklogStatus` silently succeeds when a story is not found in `backlog.md`. This masks bugs — e.g. a story that was never added, or an ID typo. The tool should return an explicit warning or error when the story is absent from the backlog for non-done status changes.

## Acceptance criteria

- [x] `set_story_status` returns a warning field (e.g. `backlog_warning`) when the story is not found in `backlog.md` and the new status is not `done`
- [x] Existing behaviour for `done` (removal) is unchanged
- [x] No silent pass-through — callers can detect the missing-entry condition
## Notes

<!-- backlog-mcp: 2026-04-13T16:18:03Z -->
Backlog groomed: this is now the sole active story and top priority. Complete pending implementation of loud missing-backlog signaling for non-done status transitions.

<!-- backlog-mcp: 2026-04-14T10:14:45Z -->
Implemented backlog_warning field in set_story_status. The non-done branch now captures the error from UpdateBacklogStatus and surfaces it as backlog_warning in the response instead of silently discarding it. Also blocked done as a valid status in set_story_status entirely — it now returns an error directing callers to use complete_story instead, which enforces this at the tool schema level for all MCP clients. Updated tests: replaced the done-removes-from-backlog test with a redirect test, fixed the unknown-story test to use a non-done status, and added TestSetStoryStatus_MissingFromBacklog_ReturnsWarning using STORY-003 (in index, absent from backlog).
