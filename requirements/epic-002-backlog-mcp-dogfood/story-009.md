# STORY-009: Gate complete_story on AC completeness

## Goal

Modify complete_story to validate acceptance criteria before allowing a story to be marked done. If unchecked items exist, require an incomplete_items parameter explaining each. If the AC section still contains only the placeholder, block completion entirely. This enforces that agents cannot skip the AC review step.

## Acceptance criteria

- [x] complete_story reads the story file and parses the ## Acceptance criteria section before proceeding
- [x] If the AC section contains only the placeholder line ('Define acceptance criteria'), completion is blocked with a clear error
- [x] If real criteria exist and all are checked (- [x]), completion proceeds as normal with no extra params required
- [x] If real criteria exist but some are unchecked (- [ ]), an incomplete_items parameter (array of strings) is required — one explanation per unchecked item
- [x] If incomplete_items is required but not provided, completion is blocked with a descriptive error listing the unchecked criteria
- [x] The incomplete_items explanations are included in the timestamped note appended to the story file
- [x] All existing complete_story tests continue to pass
- [x] New tests cover: placeholder-only AC blocks completion, all-checked proceeds, some-unchecked requires incomplete_items, incomplete_items included in note

## Notes

<!-- backlog-mcp: 2026-04-14T08:46:13Z -->
Implemented AC gating on complete_story. Added ParseAcceptanceCriteria to parser/story.go returning []ACItem. complete_story now blocks on placeholder-only AC, requires incomplete_items (one per unchecked criterion) when some criteria are unchecked, and includes those explanations in the appended note. Added optionalStringSlice helper. All 8 new tests pass alongside existing suite. Note: criteria remain unchecked in this file because check_acceptance_criterion (STORY-010) is not yet implemented — that tool is the mechanism for marking individual items [x].
