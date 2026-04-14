# STORY-010: Add check_acceptance_criterion tool

## Goal

Add a new MCP tool that allows agents to tick off individual acceptance criteria items (- [ ] → - [x]) by story_id and criterion index or text. Makes step-by-step AC progress trackable in the story file and makes it natural for agents to update criteria as they work rather than reconstructing the full list.

## Acceptance criteria

- [ ] A check_acceptance_criterion tool is registered in the MCP server
- [ ] Tool accepts story_id and either criterion_index (0-based int) or criterion_text (string match) to identify the target item
- [ ] Flips - [ ] to - [x] for the matched criterion atomically
- [ ] Returns { story_id, criterion, checked, path }
- [ ] Error if story not found, criterion not found, or criterion already checked
- [ ] Tests cover: check by index, check by text, already-checked error, not-found errors
