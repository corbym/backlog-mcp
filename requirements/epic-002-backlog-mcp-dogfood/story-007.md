# STORY-007: Add complete_story tool to enforce done lifecycle

## Goal

Agents currently must remember to call both `set_story_status` (done) and `add_story_note` at the end of a story. In practice this is easy to skip, leaving stories stuck in-progress. A dedicated `complete_story` tool makes the correct behaviour the path of least resistance and gives AGENTS.md a single, unambiguous instruction to follow.

The tool should:
- Accept `story_id` and a mandatory `summary` (the completion note)
- Atomically set status to `done` and append the summary as a timestamped note
- Return `{ story_id, completed_at, backlog_removed }` so callers can confirm the outcome

AGENTS.md should be updated to instruct agents to call `complete_story` (not `set_story_status` + `add_story_note` separately) when finishing work on a story.

## Acceptance criteria

- [ ] A `complete_story` MCP tool is registered and available to agents
- [ ] The tool accepts `story_id` (required) and `summary` (required) — calling it without `summary` returns an error, never silently omits the note
- [ ] Calling `complete_story` atomically: sets the story status to `done` in `requirements-index.md` and `backlog.md`, and appends `summary` as a timestamped note to the story file
- [ ] The tool returns `{ story_id, completed_at, backlog_removed }` on success
- [ ] Calling `complete_story` on an unknown story ID returns a clear error
- [ ] Calling `complete_story` on a story already marked `done` returns a clear error (prevents double-completion)
- [ ] `AGENTS.md` is updated: agents are instructed to call `complete_story` when finishing a story, replacing the two-step `set_story_status` + `add_story_note` pattern
- [ ] An outside-in test exercises the full slice: tool call → status updated on disk → note appended on disk → correct JSON returned
- [ ] A test asserts that omitting `summary` returns a tool error (not a panic or silent no-op)
