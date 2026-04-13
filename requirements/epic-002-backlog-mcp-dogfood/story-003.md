# STORY-003: Add create_epic MCP tool

## Goal

Agents cannot create epics via the MCP server. Add a `create_epic` tool so epics can be provisioned without touching the filesystem directly.

## Acceptance criteria

- [x] `create_epic` tool accepts `title` and optional `description`
- [x] Assigns the next available `EPIC-NNN` ID
- [x] Creates the epic directory (`epic-NNN-slug/`) and epic `.md` file
- [x] Appends the epic section to `requirements-index.md`
- [x] Returns `{ epic_id, path }`

## Notes
<!-- backlog-mcp: 2026-04-13T00:00:00Z -->
Implemented `create_epic` tool in tools.go wiring parser.CreateEpic (updated to accept description and write epic-NNN.md). Added tests: TestCreateEpic_CreatesDirectoryAndEpicFile, TestCreateEpic_AssignsNextID, TestCreateEpic_SlugInPath. All 22 tests pass.
