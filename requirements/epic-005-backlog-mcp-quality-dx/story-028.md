# STORY-028: Document bulk_update_stories criteria key format

**Type:** chore

## Goal

The `criteria` field in `bulk_update_stories` update objects expects a map of criterion text to boolean checked state, but this is not documented clearly enough for agents to use it without guessing. Add a concrete example to the tool description and AGENTS.md. Two fallback failures recorded due to this gap.

## Acceptance criteria

- [ ] Define acceptance criteria
