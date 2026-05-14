# STORY-030: Make UpdateBacklogStatus loud when story is not in backlog

**Type:** bug

## Goal

`UpdateBacklogStatus` currently succeeds silently when the target story has no entry in `backlog.md`. This masks bugs where a story's backlog entry is missing or malformed. Return a distinct error or warning so callers know the update was a no-op.

## Acceptance criteria

- [ ] Define acceptance criteria
