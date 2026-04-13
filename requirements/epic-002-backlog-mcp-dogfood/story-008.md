# STORY-008: Add set_acceptance_criteria tool to define story AC

## Goal

When create_story runs it writes a placeholder acceptance criteria section. Agents currently have no sanctioned way to replace it — direct file edits are forbidden by AGENTS.md, and add_story_note puts content in the wrong section. A dedicated set_acceptance_criteria tool gives agents a clean, explicit path to define what "done" means for a story immediately after creation (or at any point before implementation begins).

The tool should:
- Accept story_id (required) and criteria (required) — a list of strings, each becoming a `- [ ] ...` line
- Replace the existing acceptance criteria section (including the placeholder) with the new criteria
- Be idempotent: calling it a second time replaces the previous AC entirely
- Return { story_id, criteria_count, path } on success

## Acceptance criteria

- [x] A `set_acceptance_criteria` MCP tool is registered and available to agents
- [x] The tool accepts `story_id` (required) and `criteria` (required, array of strings) — omitting either returns an error
- [x] Calling the tool replaces the `## Acceptance criteria` section in the story file with one `- [ ] <criterion>` line per entry in `criteria`
- [x] The placeholder line `- [ ] Define acceptance criteria` is replaced, not duplicated
- [x] The tool is idempotent: calling it a second time replaces the previous AC entirely, leaving no stale entries
- [x] The rest of the story file (Goal, Notes, etc.) is unchanged
- [x] Calling the tool on an unknown story ID returns a clear error
- [x] Calling the tool with an empty `criteria` list returns an error — AC cannot be set to nothing
- [x] The tool returns `{ story_id, criteria_count, path }` on success
- [x] `AGENTS.md` is updated: agents are instructed to call `set_acceptance_criteria` immediately after `create_story`, before beginning implementation
- [x] An outside-in test exercises the full slice: tool call → AC section replaced on disk → correct JSON returned
- [x] A test asserts idempotency: calling twice with different criteria leaves only the second set on disk
