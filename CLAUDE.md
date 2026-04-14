# backlog-mcp: Project Context for Claude Code

## What this is

A local MCP server written in Go that gives AI agents shared read/write access to a story-based project backlog. The server exposes nine tools that let agents create epics and stories, list and read story content, set acceptance criteria, update status, append notes, and complete stories — all backed by plain markdown files on disk.

The filesystem is always the source of truth. The MCP server is a convenience layer on top, never a replacement.

---

## The story file structure

```
requirements-index.md       # master index, source of truth for status/title
backlog.md                  # priority-ordered list of not-done stories
epic-NNN-slug/
  epic-NNN.md
  story-NNN.md
```

**Story IDs:** `STORY-NNN` (e.g. `STORY-047`)  
**Epic IDs:** `EPIC-NNN` (e.g. `EPIC-003`)  
**File path pattern:** `epic-003-enemy-system/story-009.md`

**Status values:** `draft`, `in-progress`, `done`, `blocked`

Status is tracked in `requirements-index.md` (table column) and `backlog.md` (inline marker like `*(in-progress)*`). Status is NOT in the story files themselves.

**Backlog rules:**
- When a story reaches `done`, remove it from the backlog and renumber
- Priority order is head-of-list first
- Never insert in the middle — new stories append to the end

---

## Repository layout

```
backlog/
  main.go           # entry point: init subcommand + transport switch
  init.go           # scaffold a new backlog directory
  server.go         # MCP server construction, stdio/HTTP runners
  tools.go          # all 9 tool handlers
  config.go         # config loading, defaults to ./requirements
  parser/
    index.go        # parse + mutate requirements-index.md
    backlog.go      # parse + mutate backlog.md
    story.go        # filesystem scan, read story files, append notes
    atomic.go       # shared atomic write helper (temp file + rename)
  go.mod
  README.md
```

---

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BACKLOG_ROOT` | no | `requirements` | Override the path to the requirements directory |
| `BACKLOG_TRANSPORT` | no | `stdio` | Set to `http` for HTTP/SSE mode |
| `BACKLOG_HTTP_ADDR` | no | `0.0.0.0:8080` | Listen address for HTTP mode |

---

## Build and run

```bash
go mod tidy
go build -o backlog-mcp .

# Create a new backlog (first time):
./backlog-mcp init /path/to/my/backlog

# Start the server (from the project root):
./backlog-mcp    # stdio mode (default)
```

---

## The 9 tools

**`list_stories`**  
Optional filters: `epic_id` (e.g. `EPIC-003`), `status` (e.g. `draft`).  
Returns: array of `{ story_id, title, status, epic_id }`.  
Source: `requirements-index.md`.

**`get_story`**  
Required: `story_id`.  
Returns: `{ story_id, title, status, epic_id, path, content }` where `content` is the full raw markdown.  
Path resolved by filesystem scan (glob `epic-*/story-NNN.md`), metadata from index.

**`set_story_status`**  
Required: `story_id`, `status`.  
Writes atomically to:
1. `requirements-index.md` — updates status cell in the story's table row
2. `backlog.md` — if `done`, removes the entry and renumbers; otherwise updates the inline status marker

Returns: `{ story_id, old_status, new_status, backlog_removed, backlog_updated }`.  
Prefer `complete_story` over this tool when finishing a story.

**`add_story_note`**  
Required: `story_id`, `note`.  
Appends to the story `.md` file:
```
## Notes
<!-- backlog-mcp: 2026-04-13T10:32:00Z -->
<note text>
```
Returns: `{ story_id, appended_at, path }`.

**`get_index_summary`**  
No inputs.  
Returns: array of `{ epic_id, title, status, counts: {status: n}, stories: [{story_id, status}] }`.

**`create_epic`**  
Required: `title`. Optional: `description`.  
Assigns the next `EPIC-NNN` ID, creates the epic directory and `epic.md`, and registers it in `requirements-index.md`.  
Returns: `{ epic_id, path }`.

**`create_story`**  
Required: `epic_id`, `title`. Optional: `description`.  
Assigns the next `STORY-NNN` ID, writes the story file, and registers it in `requirements-index.md` and `backlog.md` with status `draft`. Story is appended to the end of the backlog.  
Returns: `{ story_id, path }`.

**`set_acceptance_criteria`**  
Required: `story_id`, `criteria` (array of strings).  
Replaces the acceptance criteria section of the story file. Each string becomes a `- [ ] ...` checklist line. Idempotent. Acceptance criteria must be set before `complete_story` can be called.  
Returns: `{ story_id, criteria_count, path }`.

**`complete_story`**  
Required: `story_id`, `summary`. Optional: `incomplete_items` (array of strings).  
Marks a story done and appends a completion summary note in one atomic call. Validates that acceptance criteria have been set and are not the default placeholder. If unchecked criteria exist, `incomplete_items` is required (one explanation per unchecked item, in order).  
Returns: `{ story_id, completed_at, backlog_removed }`.

---

## Design decisions

- **Filesystem scan for paths** — `FindStoryPath` globs `epic-*/story-NNN.md` under the requirements root. More resilient to index drift.
- **Atomic writes everywhere** — all mutations go through `writeAtomic` (temp file + rename). A crash mid-write cannot corrupt the source of truth.
- **File locking** — all mutating tools acquire a file lock (`parser.AcquireLock`) with a 5-second timeout before writing. Platform-specific implementations in `parser/lock_unix.go` and `parser/lock_windows.go`.
- **Status not in story files** — status lives only in index and backlog. Story files are append-only from the MCP's perspective (`add_story_note`).

---

/## Known issues / things to review

1. `backlog.go` regex assumes exact format `**STORY-NNN** — description`. Any deviation in the actual file will cause entries to be passed through unmodified (silently). Verify against a real backlog entry before trusting it.
2. `UpdateBacklogStatus` (non-done status changes) silently succeeds if the story isn't in the backlog. Consider making this louder.
3. `server.go` `WithBaseURL` uses `http://` hardcoded — fix before exposing over HTTPS.
