# STORY-014: List an epic's stories in the epic markdown file

**Type:** feature

## Goal

Currently epic.md files contain only the epic title and description. Users cannot see which stories belong to the epic without consulting requirements-index.md. The epic file should maintain a story list section so users can navigate directly to any story from the epic.

## Acceptance criteria

- [ ] When a story is created under an epic (via create_story), the epic.md file is updated to include a relative markdown link to the new story file (e.g. `- [STORY-014 — Story title](story-014.md)`)
- [ ] The story list in epic.md is kept in story-ID order
- [ ] Existing epic.md files without a story list section have the section appended on the next create_story call without corrupting existing content
- [ ] The story list section heading is consistent and machine-readable so future tooling can locate and update it reliably
- [ ] complete_story (or set_story_status to done) updates the story's entry in the epic list to indicate it is done (e.g. a `[x]` marker or strikethrough)
