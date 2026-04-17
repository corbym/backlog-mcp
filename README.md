[![backlog-mcp MCP server](https://glama.ai/mcp/servers/corbym/backlog-mcp/badges/score.svg)](https://glama.ai/mcp/servers/corbym/backlog-mcp)

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
go install github.com/corbym/backlog-mcp@latest
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

### Create a plan file

```bash
./backlog-mcp plan [name]
```

Creates a new plan scaffold in the `requirements/` directory. Without a name the file is `plan.md`; with a name it is `plan-<name>.md`. If the file already exists a numeric suffix is added (`plan-002.md`, etc.). Open the file and work with your agent to fill it in before creating stories.

### Configuring your MCP client

Prefer a **local** config file committed to your project root. This scopes the server to the project and means any agent cloning the repo gets the right setup automatically. Only use a global config if you want backlog-mcp available in every project without per-project configuration.

**VS Code / GitHub Copilot** — add `.mcp.json` to your project root:
```json
{
  "mcpServers": {
    "backlog-mcp": {
      "command": "/path/to/backlog-mcp"
    }
  }
}
```

**Claude Code** — add `.claude/settings.json` to your project root:
```json
{
  "mcpServers": {
    "backlog-mcp": {
      "command": "/path/to/backlog-mcp"
    }
  }
}
```

For a global fallback (applies to every project), place the same config in `~/.claude/settings.json` (Claude Code) or add it to VS Code's user `settings.json` under the `mcp.servers` key (GitHub Copilot). Always prefer the local per-project file.

---

## Tools

| Tool | Description |
|------|-------------|
| `list_stories` | List stories, optionally filtered by `epic_id` or `status` |
| `get_story` | Get full markdown content and metadata for a story |
| `get_index_summary` | High-level epic/story counts by status |
| `create_epic` | Create a new epic — assigns next EPIC-NNN ID, writes epic file, registers in index |
| `create_story` | Create a new story under an epic — assigns next STORY-NNN ID, registers in index and backlog |
| `set_epic_status` | Update epic lifecycle status with completion and regression guards (see below) |
| `set_story_status` | Update story status in index and backlog |
| `set_acceptance_criteria` | Replace the acceptance criteria section of a story (idempotent) |
| `check_acceptance_criterion` | Tick a single acceptance criterion `[x]` by index or text |
| `add_story_note` | Append a timestamped note to a story file |
| `complete_story` | Mark a story done with a mandatory completion summary and acceptance criteria validation |
| `groom_epic` | Review an epic's stories, surface gaps, and suggest missing work |

### `set_epic_status` guards

Setting status to `done` requires:

1. **`summary`** — a completion note, appended as a timestamped entry to the epic file.
2. **All stories done** — if any stories are still open the call fails and lists them. Pass `override_incomplete=true` only after the user explicitly confirms the incomplete stories are intentionally omitted.

Moving **backwards** (e.g. `done → in-progress`, `in-progress → draft`) triggers a regression prompt: the agent should offer to create new stories before proceeding. Pass `confirm_regression=true` only if the user explicitly insists on skipping that step. `blocked` and `deferred` are lateral states and can be set freely.

### `complete_story` guards

Acceptance criteria must be set (not the default placeholder) before a story can be completed. Unchecked criteria block completion unless `incomplete_items` is provided with one explanation per unchecked item. Tick done criteria `[x]` via `set_acceptance_criteria` first — do not use `incomplete_items` to confirm work that is actually finished.

---

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BACKLOG_ROOT` | no | `requirements` | Override the path to the requirements directory |

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

**Status values:** `draft`, `in-progress`, `done`, `blocked`, `deferred`

---

## Notes

- File writes are atomic (temp file + rename) — a crash mid-write cannot corrupt your files.
- The filesystem is the source of truth. The MCP server never owns the data.
