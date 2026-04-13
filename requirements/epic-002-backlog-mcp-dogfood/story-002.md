# STORY-002: Add create_story MCP tool

## Goal

Agents currently cannot create new stories via the MCP server — they must fall back to CLI or manual file editing. Add a `create_story` tool so the full story lifecycle can be managed through the MCP interface.

## Acceptance criteria

- [x] `create_story` tool accepts `title`, `epic_id`, and optional `description`
- [x] Assigns the next available `STORY-NNN` ID
- [x] Creates the story `.md` file in the correct epic directory
- [x] Appends the story to `requirements-index.md`
- [x] Appends the story to `backlog.md` with status `draft`
- [x] Returns `{ story_id, path }`

## Notes
<!-- backlog-mcp: 2026-04-13T00:00:00Z -->
Implemented `create_story` tool in tools.go wiring parser.CreateStory. Added tests: TestCreateStory_CreatesFileAndRegisters, TestCreateStory_AssignsNextID, TestCreateStory_EpicIDCaseInsensitive, TestCreateStory_UnknownEpic_ReturnsError. All 22 tests pass.
