# STORY-013: Add relative links from backlog.md entries to story and epic files

**Type:** feature

## Goal

Currently backlog.md lists stories as plain text entries. Users cannot navigate from the backlog to the relevant story file or its parent epic file. Each entry should include relative markdown links to both the story file and the epic file so that users can click through directly from the backlog.

## Acceptance criteria

- [ ] When a story is added to backlog.md (via create_story), its entry includes a relative markdown link to the story file (e.g. `[STORY-013](epic-002-backlog-mcp-dogfood/story-013.md)`)
- [ ] The backlog entry also includes a relative markdown link to the parent epic file (e.g. `[EPIC-002](epic-002-backlog-mcp-dogfood/epic-002.md)`)
- [ ] Existing entries written before this change are not broken by the new format
- [ ] UpdateBacklogStatus and RemoveFromBacklog preserve the links when modifying existing entries
- [ ] The links use paths relative to the backlog root so they resolve correctly in any markdown viewer
