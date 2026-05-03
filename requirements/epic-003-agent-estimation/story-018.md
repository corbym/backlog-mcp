# STORY-018: Surface readiness flags when a story is started

**Type:** feature

## Goal

When an agent calls set_story_status in-progress, run the readiness checklist defined in STORY-016 against the story content and include any flags in the tool response. The in-progress transition is never blocked — flags are advisory. An agent that sees flags is expected to acknowledge or address them before writing code.

## Acceptance criteria

- [ ] set_story_status in-progress response includes a readiness_flags array (empty if no flags detected)
- [ ] Flags checked: AC missing or placeholder, AC count is zero, story has no description, story notes contain unresolved questions (lines containing "?")
- [ ] Each flag in the response has a code (e.g. "NO_AC", "VAGUE_AC") and a human-readable message
- [ ] Response includes a ready boolean: true when readiness_flags is empty, false otherwise
- [ ] The in-progress transition always succeeds regardless of flags
- [ ] Tests cover: clean story (ready: true, empty flags), story with missing AC, story with placeholder AC, story with open questions in notes
