# Agent Instructions

## Backlog (backlog-mcp)

Project stories and epics live in `requirements/`. Use the backlog-mcp MCP tools to interact with them — do not edit the index or backlog files by hand.

**Tools**

- `list_stories` — list stories, optionally filtered by `epic_id` or `status`
- `get_story` — read the full markdown content and metadata for a story
- `set_story_status` — update a story's status (`draft`, `in-progress`, `done`, `blocked`)
- `add_story_note` — append a timestamped note to a story file
- `get_index_summary` — high-level epic/story counts by status

**Conventions**

- Story IDs: `STORY-NNN` (e.g. `STORY-047`)
- Epic IDs: `EPIC-NNN` (e.g. `EPIC-003`)
- Always call `set_story_status` when you start or finish work on a story
- Use `add_story_note` to record decisions, blockers, or progress — never edit story files directly

**Plans**

Plans live in `requirements/plan*.md`. They describe overall project goals and direction.

- Read all plan files before creating or prioritising stories
- When you find a plan marked `Status: draft`, ask the user the questions in each section and fill it in before proceeding
- Use plans to decide what stories to create and how to order the backlog
- Plans never reference specific stories — keep that relationship one-way (stories may reference plan sections, not the other way around)
- Never edit a plan to add story references or implementation detail — plans are for goals and intent, not execution tracking
