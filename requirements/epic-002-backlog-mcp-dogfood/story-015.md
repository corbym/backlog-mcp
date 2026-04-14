# STORY-015: Add groom_epic tool to reconcile epic Stories section with filesystem and index

**Type:** feature

## Goal

The ## Stories section in an epic.md file is only written when create_story is called. Any drift — stories added outside the tool, deleted story files, or title/status changes — goes undetected. A groom_epic tool should scan the filesystem for all story-NNN.md files under the epic directory, compare against the ## Stories section, and reconcile: add missing entries, remove entries whose files no longer exist, and refresh titles and done/undone status from the index.

## Acceptance criteria

- [x] groom_epic accepts an epic_id and scans all story-NNN.md files under that epic's directory on the filesystem
- [x] Any story file present on disk but missing from the ## Stories section is added as a new entry (status and title sourced from requirements-index.md; falls back to reading the story file's first heading if not in the index)
- [x] Any entry in the ## Stories section whose story file no longer exists on disk is removed
- [x] Existing entries have their title refreshed from the index if it has changed
- [x] Existing entries have their done/undone marker (- [x] vs - [ ]) refreshed to match the status in requirements-index.md
- [x] Entries in the ## Stories section remain in ascending story-ID order after grooming
- [x] If the epic.md has no ## Stories section and there are story files on disk, the section is created
- [x] groom_epic returns a summary: {epic_id, added: [story_ids], removed: [story_ids], updated: [story_ids], unchanged: [story_ids]}

## Notes

<!-- backlog-mcp: 2026-04-14T10:38:09Z -->
Implemented groom_epic tool in parser/groom.go and wired into tools.go. The tool scans story-NNN.md files on disk, reconciles the ## Stories section in epic.md, and returns {epic_id, added, removed, updated, unchanged}. Verified working: grooming EPIC-001, EPIC-002, and EPIC-003 all ran successfully in this session, adding all missing story entries to epic files.
