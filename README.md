# backlog-mcp

An MCP server that gives AI agents structured read/write access to a story-based project backlog. Agents can list stories, read content, update status, and append notes — all backed by plain markdown files that live inside your project repository.

## How collaboration works

There is no shared server. The backlog files live in your repo under `requirements/`, committed and versioned alongside your code. Collaboration between agents, or between an agent and a human, works exactly the way the rest of your codebase does: through git. If two agents update different stories concurrently, git merges them. If they touch the same line, you resolve it like any other merge conflict.

The MCP server is a local process each agent runs for itself. It reads and writes files; git handles the rest.

---

## Install

Download the latest binary for your platform from the [Releases](../../releases) page and put it somewhere on your `$PATH`.

Or, if you have Go installed:

```bash
go install backlog@latest
```

---

## Build from source

```bash
go mod tidy
go build -o backlog-mcp .
```

---

## Setup

Initialise a `requirements/` folder in your project root:

```bash
./backlog-mcp init /path/to/your/project/requirements
```

This creates:

```
requirements/
  requirements-index.md   # master index — source of truth for epics and story status
  backlog.md              # priority-ordered list of not-done stories
  epic-001-example/
    story-001.md          # example story file
```

Commit the `requirements/` folder to your repo. Edit the files to add your own epics and stories.

---

## Running

```bash
./backlog-mcp
```

The server looks for a `requirements/` directory relative to the working directory it is launched from. Claude Code sets the working directory to the project root, so no configuration is needed.

**Claude Code config** (`.claude/settings.json` in your project, or `~/.claude/settings.json` globally):
```json
{
  "mcpServers": {
    "backlog-mcp": {
      "command": "/path/to/backlog-mcp"
    }
  }
}
```

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
| `BACKLOG_ROOT` | no | `requirements` | Override the path to the requirements directory |
| `BACKLOG_TRANSPORT` | no | `stdio` | Set to `http` for HTTP/SSE mode |
| `BACKLOG_HTTP_ADDR` | no | `0.0.0.0:8080` | Listen address for HTTP mode |

---

## File format

**`requirements-index.md`** — one epic section per heading, one story per table row:

```markdown
## EPIC-001: Combat System — `draft`

| Story | Title | Status |
|-------|-------|--------|
| [STORY-001](./epic-001-combat-system/story-001.md) | Basic combat | draft |
```

**`backlog.md`** — priority-ordered numbered list:

```markdown
1. **STORY-001** — Basic combat
2. **STORY-002** — Enemy AI *(in-progress)*
```

**Story files** live at `epic-NNN-slug/story-NNN.md` under `BACKLOG_ROOT`.

**Status values:** `draft`, `in-progress`, `done`, `blocked`

---

## Notes

- File writes are atomic (temp file + rename) — a crash mid-write cannot corrupt your files.
- The filesystem is the source of truth. The MCP server never owns the data.
