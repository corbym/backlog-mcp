# Agent Instructions

## Backlog (backlog-mcp)

Project stories and epics live in `requirements/`. Use the backlog-mcp MCP tools to interact with them — do not edit the index or backlog files by hand.

**Tools**

- `list_stories` — list stories, optionally filtered by `epic_id` or `status`
- `get_story` — read the full markdown content and metadata for a story
- `set_story_status` — update a story's status (`draft`, `in-progress`, `done`, `blocked`)
- `add_story_note` — append a timestamped note to a story file
- `complete_story` — mark a story `done` and append the completion summary note in one call
- `get_index_summary` — high-level epic/story counts by status

**Conventions**

- Story IDs: `STORY-NNN` (e.g. `STORY-047`)
- Epic IDs: `EPIC-NNN` (e.g. `EPIC-003`)
- Always call `set_story_status` when you start work on a story (`in-progress`)
- Always call `complete_story` when you finish work on a story (do not use `set_story_status` + `add_story_note` separately)
- Use `add_story_note` to record decisions, blockers, or progress — never edit story files directly
- After `create_story`, immediately call `set_acceptance_criteria` with a concrete list of criteria before beginning implementation — a story with only the placeholder AC is not ready to implement

**Testing (mandatory)**

- **Always write tests first.** No production code is written before a failing test exists that demands it.
- **Always write outside-in.** Start from the outermost entry point (tool handler, HTTP handler, public API) and work inward. The first test must exercise the full slice of behaviour from the outside; only then add lower-level unit tests as needed to drive internal design.
- The test must fail for the right reason before any implementation is written. Verify the failure message makes sense.
- Only write enough production code to make the current failing test pass, then refactor before moving on.
- Do not write tests after the fact to cover code that already exists — if you find yourself doing this, stop, delete the code, and restart test-first.

**Plans**

Plans live in `requirements/plan*.md`. They describe overall project goals and direction.

- Read all plan files before creating or prioritising stories
- When you find a plan marked `Status: draft`, ask the user the questions in each section and fill it in before proceeding
- Use plans to decide what stories to create and how to order the backlog
- Plans never reference specific stories — keep that relationship one-way (stories may reference plan sections, not the other way around)
- Never edit a plan to add story references or implementation detail — plans are for goals and intent, not execution tracking

## Bulk update tools

- Use `bulk_update_acceptance_criteria` when ticking off two or more criteria on the same story in one operation
- Use `bulk_update_stories` when a single PR or task affects multiple stories, for example updating status and appending a note across three stories at once
- Use `bulk_update_epics` only when explicitly instructed to update epic status or notes, never infer epic completion from story statuses
- Prefer bulk tools over repeated single-item calls whenever two or more items need the same class of update
- Never use bulk tools to create or delete stories or epics, only to update existing ones

## PR Agent Behaviour

- The agent runs automatically on PRs via GitHub Actions (`.github/workflows/backlog-agent.yml`), triggered on `opened` and `synchronize` events
- It reads the PR title, description, and branch name to infer which stories are affected — explicit story IDs (e.g. `STORY-042`) in the title or branch name are the most reliable signal; without them the agent infers from context
- It may update story status to `in-progress` if the PR is opened and the story is currently `draft`
- It may append a note to the relevant story with the PR number and a one-line summary of the PR's purpose
- It must never set story status to `done` on an open PR — `done` is only set when work is merged and complete
- It must never create or delete stories, only update existing ones
