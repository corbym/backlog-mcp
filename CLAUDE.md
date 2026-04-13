# backlog: Project Context for Claude Code

## What this is

A local MCP server written in Go that gives AI agents (Claude Code, GitHub Copilot) shared read/write access to the story management files in the Deep project. The key idea: both agents work on the same project but have no shared memory — this server is the coordination layer.

The file system is always the source of truth. The MCP server is a convenience layer on top, never a replacement.

---

## The existing story file structure

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

**Backlog rules (from the file itself):**
- When a story reaches `done`, remove it from the backlog and renumber
- Priority order is head-of-list first
- Never insert in the middle — new stories append to the end

---

## The MCP server: deep-mcp

### Repository layout

```
deep-mcp/
  main.go           # entry point, transport switch
  server.go         # MCP server construction, stdio/HTTP runners
  tools.go          # all 5 tool handlers
  config.go         # DEEP_STORIES_ROOT env var loading
  parser/
    index.go        # parse + mutate requirements-index.md
    backlog.go      # parse + mutate backlog.md
    story.go        # filesystem scan, read story files, append notes
    atomic.go       # shared atomic write helper (temp file + rename)
  go.mod
  README.md
```

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DEEP_STORIES_ROOT` | yes | — | Path to the directory containing `requirements-index.md` |
| `DEEP_TRANSPORT` | no | `stdio` | Set to `http` for HTTP/SSE mode |
| `DEEP_HTTP_ADDR` | no | `0.0.0.0:8080` | Listen address for HTTP mode |

### Build and run

```bash
cd deep-mcp
go mod tidy
go build -o deep-mcp .

export DEEP_STORIES_ROOT=/path/to/deep/stories
./deep-mcp    # stdio mode (default)
```

### The 5 tools

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
<!-- DEEP-MCP: 2026-04-13T10:32:00Z -->
<note text>
```
Returns: `{ story_id, appended_at, path }`.

**`get_index_summary`**  
No inputs.  
Returns: array of `{ epic_id, title, status, counts: {status: n}, stories: [{story_id, status}] }`.  
Good for situational awareness without reading every file.

---

## Design decisions

- **Filesystem scan for paths** — index links are not parsed for file paths; `FindStoryPath` globs `epic-*/story-NNN.md` under `DEEP_STORIES_ROOT`. More resilient to index drift.
- **Atomic writes everywhere** — all mutations go through `writeAtomic` (temp file + rename). A crash mid-write cannot corrupt the source of truth.
- **No locking in PoC** — single-writer assumption. Add a file lock before allowing concurrent agent writes.
- **Status not in story files** — status lives only in index and backlog. Story files are append-only from the MCP's perspective (`add_story_note`).

---

## PoC phases

**PoC-A (local, stdio):** Claude Code connects via stdio. Validates tool surface and file I/O on real stories.

**PoC-B (hosted, HTTP/SSE):** Server deployed to a small instance (Fly.io / t3.micro), story files synced from repo. Validates GitHub Copilot (github.com) can connect — this is the critical unknown.

For PoC-B: set `DEEP_TRANSPORT=http`. Note the `WithBaseURL` in `server.go` is hardcoded to `http://` — change to `https://` and put TLS termination in front (nginx/load balancer) before exposing publicly.

---

## Known issues / things to review

1. `backlog.go` regex assumes exact format `**STORY-NNN** — description`. Any deviation in the actual file will cause entries to be passed through unmodified (silently). Verify against a real backlog entry before trusting it.
2. `UpdateBacklogStatus` (non-done status changes) silently succeeds if the story isn't in the backlog. Consider making this louder.
3. `server.go` `WithBaseURL` uses `http://` hardcoded — fix before PoC-B.