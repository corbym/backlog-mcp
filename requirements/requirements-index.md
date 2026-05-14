# Requirements Index

## EPIC-001: Example Epic — `done`

| Story | Title | Status |
|-------|-------|--------|
| [STORY-001](./epic-001-example/story-001.md) | Example story | done |

## EPIC-002: Backlog MCP Dogfood — `done`

| Story | Title | Status |
|-------|-------|--------|
| [STORY-002](./epic-002-backlog-mcp-dogfood/story-002.md) | Add create_story MCP tool | done |
| [STORY-003](./epic-002-backlog-mcp-dogfood/story-003.md) | Add create_epic MCP tool | done |
| [STORY-004](./epic-002-backlog-mcp-dogfood/story-004.md) | Add file locking for concurrent agent writes | done |
| [STORY-005](./epic-002-backlog-mcp-dogfood/story-005.md) | Fix hardcoded http:// base URL in server.go | done |
| [STORY-006](./epic-002-backlog-mcp-dogfood/story-006.md) | Surface loud failure when set_story_status target is not in backlog | done |
| [STORY-007](./epic-002-backlog-mcp-dogfood/story-007.md) | Add complete_story tool to enforce done lifecycle | done |
| [STORY-008](./epic-002-backlog-mcp-dogfood/story-008.md) | Add set_acceptance_criteria tool to define story AC | done |
| [STORY-009](./epic-002-backlog-mcp-dogfood/story-009.md) | Gate complete_story on AC completeness | done |
| [STORY-010](./epic-002-backlog-mcp-dogfood/story-010.md) | Add check_acceptance_criterion tool | done |
| [STORY-011](./epic-002-backlog-mcp-dogfood/story-011.md) | Document all 9 MCP tools and fix set_acceptance_criteria schema bug | done |
| [STORY-012](./epic-002-backlog-mcp-dogfood/story-012.md) | Add story_type field to create_story tool | done |
| [STORY-013](./epic-002-backlog-mcp-dogfood/story-013.md) | Add relative links from backlog.md entries to story and epic files | done | feature |
| [STORY-014](./epic-002-backlog-mcp-dogfood/story-014.md) | List an epic's stories in the epic markdown file | done | feature |
| [STORY-015](./epic-002-backlog-mcp-dogfood/story-015.md) | Add groom_epic tool to reconcile epic Stories section with filesystem and index | done | feature |
| [STORY-021](./epic-002-backlog-mcp-dogfood/story-021.md) | Add set_epic_status tool to manage epic lifecycle | done | feature |
| [STORY-022](./epic-002-backlog-mcp-dogfood/story-022.md) | Add bulk update tools for stories, epics, and acceptance criteria | done | feature |

## EPIC-003: Agent Estimation — `draft`

| Story | Title | Status |
|-------|-------|--------|
| [STORY-016](./epic-003-agent-estimation/story-016.md) | Spike: define estimation dimensions and scoring rubric | draft | spike |
| [STORY-017](./epic-003-agent-estimation/story-017.md) | Capture actuals on story completion | deferred | feature |
| [STORY-018](./epic-003-agent-estimation/story-018.md) | Agent pre-flight self-estimate recorded on story start | draft | feature |
| [STORY-019](./epic-003-agent-estimation/story-019.md) | Calibration report: compare pre-flight estimates vs actuals | deferred | feature |
| [STORY-020](./epic-003-agent-estimation/story-020.md) | Automate pre-flight scoring on set_story_status in-progress | deferred | feature |

## EPIC-004: Distribution — `in-progress`

| Story | Title | Status |
|-------|-------|--------|
| [STORY-023](./epic-004-distribution/story-023.md) | Create Homebrew tap formula and release automation | in-progress |

## EPIC-005: Backlog MCP Quality & DX — `draft`

| Story | Title | Status | Type |
|-------|-------|--------|------|
| [STORY-024](./epic-005-backlog-mcp-quality-dx/story-024.md) | Add reorder_backlog MCP tool | draft | feature |
| [STORY-025](./epic-005-backlog-mcp-quality-dx/story-025.md) | Fix em-dash encoding artefacts in acceptance criteria text | draft | bug |
| [STORY-026](./epic-005-backlog-mcp-quality-dx/story-026.md) | Add tool-surface discovery hint to at least one tool description | draft | chore |
| [STORY-027](./epic-005-backlog-mcp-quality-dx/story-027.md) | Update AGENTS.md with complete tool surface | deferred | chore |
| [STORY-028](./epic-005-backlog-mcp-quality-dx/story-028.md) | Document bulk_update_stories criteria key format | draft | chore |
| [STORY-029](./epic-005-backlog-mcp-quality-dx/story-029.md) | Fix backlog.go regex to handle format deviations gracefully | draft | bug |
| [STORY-030](./epic-005-backlog-mcp-quality-dx/story-030.md) | Make UpdateBacklogStatus loud when story is not in backlog | draft | bug |
