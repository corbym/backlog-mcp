# backlog-mcp: Project Context for Claude Code

## What this is

A local MCP server written in Go that gives AI agents shared read/write access to a story-based project backlog. The server exposes five tools that let agents list stories, read story content, update status, and append notes — all backed by plain markdown files on disk.

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
  tools.go          # all 5 tool handlers
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

## The 5 tools

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

---

## Design decisions

- **Filesystem scan for paths** — `FindStoryPath` globs `epic-*/story-NNN.md` under the requirements root. More resilient to index drift.
- **Atomic writes everywhere** — all mutations go through `writeAtomic` (temp file + rename). A crash mid-write cannot corrupt the source of truth.
- **No locking in PoC** — single-writer assumption. Add a file lock before allowing concurrent agent writes.
- **Status not in story files** — status lives only in index and backlog. Story files are append-only from the MCP's perspective (`add_story_note`).

---

/## Known issues / things to review

1. `backlog.go` regex assumes exact format `**STORY-NNN** — description`. Any deviation in the actual file will cause entries to be passed through unmodified (silently). Verify against a real backlog entry before trusting it.
2. `UpdateBacklogStatus` (non-done status changes) silently succeeds if the story isn't in the backlog. Consider making this louder.
3. `server.go` `WithBaseURL` uses `http://` hardcoded — fix before exposing over HTTPS.
