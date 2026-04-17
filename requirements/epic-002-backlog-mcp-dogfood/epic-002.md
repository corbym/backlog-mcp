# EPIC-002: Backlog MCP Dogfood

## Goal

Build out the full set of MCP tools needed to manage this project's own backlog through the MCP server — dogfooding the tool on its own development. Covers the complete story lifecycle: creating epics and stories, setting acceptance criteria, updating status, appending notes, and completing stories with a summary. Includes file locking for concurrent agent safety, loud failure modes for edge cases, and tooling quality improvements surfaced through real usage.

## Stories

- [x] [STORY-002](story-002.md) — Add create_story MCP tool
- [x] [STORY-003](story-003.md) — Add create_epic MCP tool
- [x] [STORY-004](story-004.md) — Add file locking for concurrent agent writes
- [x] [STORY-005](story-005.md) — Fix hardcoded http:// base URL in server.go
- [x] [STORY-006](story-006.md) — Surface loud failure when set_story_status target is not in backlog
- [x] [STORY-007](story-007.md) — Add complete_story tool to enforce done lifecycle
- [x] [STORY-008](story-008.md) — Add set_acceptance_criteria tool to define story AC
- [x] [STORY-009](story-009.md) — Gate complete_story on AC completeness
- [x] [STORY-010](story-010.md) — Add check_acceptance_criterion tool
- [x] [STORY-011](story-011.md) — Document all 9 MCP tools and fix set_acceptance_criteria schema bug
- [x] [STORY-012](story-012.md) — Add story_type field to create_story tool
- [x] [STORY-013](story-013.md) — Add relative links from backlog.md entries to story and epic files
- [x] [STORY-014](story-014.md) — List an epic's stories in the epic markdown file
- [x] [STORY-015](story-015.md) — Add groom_epic tool to reconcile epic Stories section with filesystem and index
- [x] [STORY-021](story-021.md) — Add set_epic_status tool to manage epic lifecycle
