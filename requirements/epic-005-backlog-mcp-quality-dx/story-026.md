# STORY-026: Add tool-surface discovery hint to at least one tool description

**Type:** chore

## Goal

In VS Code Copilot (and similar deferred-tool environments), MCP tools are not loaded into agent context until `tool_search` is called. If at least one tool description (e.g. `list_stories`) includes a line listing all other tools in the server, any agent that discovers one tool via search immediately sees the full surface. Three separate incidents of agents falling back to direct file edits or missing AC tools are traceable to this discovery gap.

## Acceptance criteria

- [ ] Define acceptance criteria
