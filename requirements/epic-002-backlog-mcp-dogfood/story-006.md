# STORY-006: Surface loud failure when set_story_status target is not in backlog

## Goal

`UpdateBacklogStatus` silently succeeds when a story is not found in `backlog.md`. This masks bugs — e.g. a story that was never added, or an ID typo. The tool should return an explicit warning or error when the story is absent from the backlog for non-done status changes.

## Acceptance criteria

- [ ] `set_story_status` returns a warning field (e.g. `backlog_warning`) when the story is not found in `backlog.md` and the new status is not `done`
- [ ] Existing behaviour for `done` (removal) is unchanged
- [ ] No silent pass-through — callers can detect the missing-entry condition