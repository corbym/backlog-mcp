# STORY-021: Add set_epic_status tool to manage epic lifecycle

**Type:** feature

## Goal

The MCP server has no way to update an epic's status. Agents must manually edit requirements-index.md to close out a completed epic. Add a set_epic_status tool that updates the epic heading status marker in requirements-index.md atomically, mirroring the pattern used by set_story_status.

## Acceptance criteria

- [x] A set_epic_status MCP tool is registered and callable
- [x] Tool accepts required parameters epic_id and status
- [x] Valid status values are draft, in-progress, done, blocked, and deferred — invalid values return an error
- [x] Tool updates the epic heading status marker in requirements-index.md atomically using the existing writeAtomic helper
- [x] File lock is acquired before writing, consistent with other mutating tools
- [x] Returns { epic_id, old_status, new_status } on success
- [x] Returns a clear error if the epic_id does not exist in the index
- [x] Tool description clearly explains when to use set_epic_status versus set_story_status, and what each status value means for an epic

## Notes

<!-- backlog-mcp: 2026-04-17T14:32:36Z -->
Added set_epic_status tool to tools.go and UpdateEpicStatus to parser/index.go. The tool validates status values (draft, in-progress, done, blocked, deferred), acquires a file lock, updates the epic heading backtick-marker in requirements-index.md atomically, and returns {epic_id, old_status, new_status}. Tool description distinguishes it clearly from set_story_status. All tests pass.
