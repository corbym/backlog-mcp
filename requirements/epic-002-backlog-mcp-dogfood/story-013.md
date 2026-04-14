# STORY-013: Add relative links from backlog.md entries to story and epic files

**Type:** feature

## Goal

Currently backlog.md lists stories as plain text entries. Users cannot navigate from the backlog to the relevant story file or its parent epic file. Each entry should include relative markdown links to both the story file and the epic file so that users can click through directly from the backlog.

## Acceptance criteria

- [x] When a story is added to backlog.md (via create_story), its entry includes a relative markdown link to the story file (e.g. `[STORY-013](epic-002-backlog-mcp-dogfood/story-013.md)`)
- [x] The backlog entry also includes a relative markdown link to the parent epic file (e.g. `[EPIC-002](epic-002-backlog-mcp-dogfood/epic-002.md)`)
- [x] Existing entries written before this change are not broken by the new format
- [x] UpdateBacklogStatus and RemoveFromBacklog preserve the links when modifying existing entries
- [x] The links use paths relative to the backlog root so they resolve correctly in any markdown viewer

## Notes

<!-- backlog-mcp: 2026-04-14T10:38:05Z -->
Implemented relative markdown links in backlog.md entries. appendStoryToBacklog now writes entries in the format `N. [STORY-NNN](path) ([EPIC-NNN](path)) — title`. The backlogEntryRe regex was updated to match both old bold-text format and new link format, so UpdateBacklogStatus and RemoveFromBacklog handle both without breaking existing entries.
