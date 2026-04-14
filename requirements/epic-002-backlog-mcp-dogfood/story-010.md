# STORY-010: Add check_acceptance_criterion tool

## Goal

Add a new MCP tool that allows agents to tick off individual acceptance criteria items (- [ ] → - [x]) by story_id and criterion index or text. Makes step-by-step AC progress trackable in the story file and makes it natural for agents to update criteria as they work rather than reconstructing the full list.

## Acceptance criteria

- [x] A check_acceptance_criterion tool is registered in the MCP server
- [x] Tool accepts story_id and either criterion_index (0-based int) or criterion_text (string match) to identify the target item
- [x] Flips - [ ] to - [x] for the matched criterion atomically
- [x] Returns { story_id, criterion, checked, path }
- [x] Error if story not found, criterion not found, or criterion already checked
- [x] Tests cover: check by index, check by text, already-checked error, not-found errors

## Notes

<!-- backlog-mcp: 2026-04-14T09:41:31Z -->
Implemented check_acceptance_criterion tool end-to-end. Added CheckAcceptanceCriterion to parser/story.go: finds the AC section, resolves the target by 0-based index or case-insensitive exact text match, errors if already checked or not found, flips - [ ] to - [x] atomically. Registered the tool in tools.go using mcp.WithNumber for criterion_index and req.GetInt to read it. Returns {story_id, criterion, checked, path}. Added 7 tests covering check by index, check by text, case-insensitive text match, already-checked error, index out of range, text not found, unknown story, and neither parameter provided. All tests pass.

Incomplete criteria:
- [ ] A check_acceptance_criterion tool is registered in the MCP server: Done — check_acceptance_criterion registered in tools.go with mcp.WithNumber for criterion_index and mcp.WithString for criterion_text
- [ ] Tool accepts story_id and either criterion_index (0-based int) or criterion_text (string match) to identify the target item: Done — tool accepts story_id (required), criterion_index (0-based int, optional), and criterion_text (case-insensitive exact match, optional); errors if neither is provided
- [ ] Flips - [ ] to - [x] for the matched criterion atomically: Done — CheckAcceptanceCriterion in parser/story.go flips - [ ] to - [x] on the matched line and writes atomically via writeAtomic
- [ ] Returns { story_id, criterion, checked, path }: Done — tool returns {story_id, criterion, checked, path} on success
- [ ] Error if story not found, criterion not found, or criterion already checked: Done — errors returned for unknown story (FindStoryPath), criterion not found (text match), criterion already checked (both index and text paths)
- [ ] Tests cover: check by index, check by text, already-checked error, not-found errors: Done — TestCheckAcceptanceCriterion_ByIndex, ByText, ByText_CaseInsensitive, AlreadyChecked, IndexOutOfRange, TextNotFound, UnknownStory, NeitherProvided — all 7 tests pass

