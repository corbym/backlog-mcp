# deep-mcp

A local MCP server providing shared story-state access for AI agents working on the Deep project.

Both Claude.ai and GitHub Copilot can read and write story status through this server, giving them a shared coordination layer without needing to know about each other.

---

## Prerequisites

- Go 1.22+
- The Deep project stories directory (contains `requirements-index.md` and `backlog.md`)

---

## Setup

```bash
cd deep-mcp
go mod tidy
go build -o deep-mcp .
```

---

## Running

### PoC-A: Local (stdio) — for Claude Code / Claude.ai connector

```bash
export DEEP_STORIES_ROOT=/path/to/your/deep/stories
./deep-mcp
```

stdio transport is the default. The process communicates over stdin/stdout.

**Claude.ai connector config** (`~/.config/claude/mcp.json` or equivalent):
```json
{
  "mcpServers": {
    "deep-mcp": {
      "command": "/path/to/deep-mcp",
      "env": {
        "DEEP_STORIES_ROOT": "/path/to/your/deep/stories"
      }
    }
  }
}
```

### PoC-B: Hosted (HTTP/SSE) — for GitHub Copilot

```bash
export DEEP_STORIES_ROOT=/path/to/stories
export DEEP_TRANSPORT=http
export DEEP_HTTP_ADDR=0.0.0.0:8080
./deep-mcp
```

Point Copilot's MCP config at `http://your-server:8080/sse`.

---

## Tools

| Tool | Description |
|------|-------------|
| `list_stories` | List stories, optionally filtered by `epic_id` or `status` |
| `get_story` | Get full markdown content and metadata for a story |
| `set_story_status` | Update story status in index and backlog |
| `add_story_note` | Append a timestamped note to a story file |
| `get_index_summary` | High-level epic/story counts by status |

---

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DEEP_STORIES_ROOT` | yes | — | Path to the directory containing `requirements-index.md` |
| `DEEP_TRANSPORT` | no | `stdio` | Set to `http` for HTTP/SSE mode |
| `DEEP_HTTP_ADDR` | no | `0.0.0.0:8080` | Listen address for HTTP mode |

---

## The validation test (PoC round-trip)

1. Agent calls `list_stories` — finds a ready story
2. Agent calls `get_story` — reads the acceptance criteria
3. Agent does implementation work
4. Agent calls `add_story_note` — records what it did
5. Agent calls `set_story_status` with `done`
6. `requirements-index.md` and `backlog.md` are updated automatically
7. A second agent (or human) calls `get_story` or `get_index_summary` — sees the completed state

---

## Notes

- File writes are atomic (temp file + rename) so a crash mid-write cannot corrupt the source of truth.
- No locking in PoC — single-writer assumption. Add a file lock before multi-agent concurrent writes.
- The filesystem is the source of truth. The MCP server never owns the data.
