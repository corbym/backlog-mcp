# STORY-007: Add complete_story tool to enforce done lifecycle

## Goal

Agents currently must remember to call both `set_story_status` (done) and `add_story_note` at the end of a story. In practice this is easy to skip, leaving stories stuck in-progress. A dedicated `complete_story` tool makes the correct behaviour the path of least resistance and gives AGENTS.md a single, unambiguous instruction to follow.

The tool should:
- Accept `story_id` and a mandatory `summary` (the completion note)
- Atomically set status to `done` and append the summary as a timestamped note
- Return `{ story_id, completed_at, backlog_removed }` so callers can confirm the outcome

AGENTS.md should be updated to instruct agents to call `complete_story` (not `set_story_status` + `add_story_note` separately) when finishing work on a story.

## Acceptance criteria

- [x] A `complete_story` MCP tool is registered and available to agents
- [x] The tool accepts `story_id` (required) and `summary` (required) — calling it without `summary` returns an error, never silently omits the note
- [x] Calling `complete_story` atomically: sets the story status to `done` in `requirements-index.md` and `backlog.md`, and appends `summary` as a timestamped note to the story file
- [x] The tool returns `{ story_id, completed_at, backlog_removed }` on success
- [x] Calling `complete_story` on an unknown story ID returns a clear error
- [x] Calling `complete_story` on a story already marked `done` returns a clear error (prevents double-completion)
- [x] `AGENTS.md` is updated: agents are instructed to call `complete_story` when finishing a story, replacing the two-step `set_story_status` + `add_story_note` pattern
- [x] An outside-in test exercises the full slice: tool call → status updated on disk → note appended on disk → correct JSON returned
- [x] A test asserts that omitting `summary` returns a tool error (not a panic or silent no-op)

## Notes

<!-- backlog-mcp: 2026-04-13T15:53:27Z -->
Implemented STORY-007 end-to-end with test-first outside-in flow. Added failing complete_story tests first, implemented complete_story tool with required story_id + summary, status-to-done transition, backlog removal, and completion note append, plus clear errors for missing summary, unknown story, and already-done story. Updated AGENTS.md to instruct complete_story for finish lifecycle. Verified with `go test ./... -count=1` passing.
