# STORY-031: Fix set_acceptance_criteria corruption when passed pre-ticked checkbox strings

**Type:** bug

## Goal

set_acceptance_criteria unconditionally prepends `- [ ] ` to each criterion string, even when the string already begins with `[x] AC-ID: ...` or `- [x] ...`. This produces double-nested broken checkboxes (`- [ ] [x] AC-ID: ...`). The tool then runs assignMissingIDs over the corrupted section — the original ID is treated as part of the text content, new IDs are assigned, and the file is left with all AC IDs changed. Any subsequent tool call referencing the old IDs fails silently. Recovery requires a direct file edit.

Fix: detect and strip any leading `- [ ] ` or `- [x] ` prefix (and bare `[x]`/`[ ]`) from each input string before writing the canonical format. Preserve ticked state.

## Acceptance criteria

- [x] AC-STORY-031-b7f51bff: set_acceptance_criteria called with a string already starting with `[x] text` writes `- [x] text` (not `- [ ] [x] text`)
- [x] AC-STORY-031-d173b106: set_acceptance_criteria called with a string starting with `- [x] text` normalises to `- [x] text` preserving ticked state
- [x] AC-STORY-031-4c1ac212: set_acceptance_criteria called with a string containing an existing AC ID (`AC-STORY-NNN-xxxxxxxx: text`) preserves the ID rather than regenerating it
- [x] AC-STORY-031-f0e9b369: A failing test is written first that reproduces the double-nesting with a pre-ticked input string
- [x] AC-STORY-031-c0efa502: A round-trip fidelity test confirms no AC IDs change when set_acceptance_criteria is called with the current criteria re-read from the file
- [x] AC-STORY-031-a8176e36: All existing tests pass

## Notes

<!-- backlog-mcp: 2026-07-02T18:32:10Z -->
Added `normalizeACInput` helper to `parser/story.go` that strips leading checkbox markers (`- [x] `, `- [ ] `, `[x] `, `[ ] `) from a criterion string and returns the bare text plus ticked state. Updated the criterion-writing loop in `SetAcceptanceCriteria` to call `normalizeACInput` before `parseACText`, so pre-ticked or fully-formatted input strings are written as properly structured lines without double-nesting. Added `hasCriterionLineAtStart` test helper that checks the outer checkbox marker position (not just substring presence). Added four new tests covering bare `[x]` input, dashed `- [x]` input, full stored-line round-trip with ID preservation, and a general round-trip fidelity test. All 77 tests pass.
