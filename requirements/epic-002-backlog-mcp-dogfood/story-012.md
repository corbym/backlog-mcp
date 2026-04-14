# STORY-012: Add story_type field to create_story tool

## Goal

Add a story_type parameter (feature, bug, chore, spike) to create_story. Store the type in the story file and requirements-index.md. Surface it in list_stories and get_story responses. Add story_type filter to list_stories.

## Acceptance criteria

- [x] create_story accepts a story_type parameter with valid values: feature, bug, chore, spike
- [x] story_type defaults to 'feature' if not provided
- [x] story_type is written into the story .md file header
- [x] story_type is stored in requirements-index.md alongside the story row
- [x] list_stories returns story_type in each result object
- [x] list_stories accepts a story_type filter parameter
- [x] get_story returns story_type in its result
- [x] Build passes with no errors

## Notes

<!-- backlog-mcp: 2026-04-14T09:25:55Z -->
Added story_type field (feature, bug, chore, spike) throughout the stack. parser.Story struct gains StoryType field. storyRowRe updated to optionally capture a 4th type column — rows without it default to "feature" for backward compat. storyContent writes **Type:** metadata into new story files. appendStoryToIndex writes a 4-column row. appendEpicToIndex creates 4-column table headers for new epics. create_story tool accepts optional story_type param (defaults to "feature", validates against allowed values). list_stories now returns story_type in results and accepts a story_type filter param. get_story returns story_type. Build clean.

Incomplete criteria:
- [ ] create_story accepts a story_type parameter with valid values: feature, bug, chore, spike: Done — create_story accepts story_type with valid values feature, bug, chore, spike
- [ ] story_type defaults to 'feature' if not provided: Done — defaults to 'feature' when story_type is omitted
- [ ] story_type is written into the story .md file header: Done — **Type:** field written into story .md file header via storyContent
- [ ] story_type is stored in requirements-index.md alongside the story row: Done — 4-column row written to requirements-index.md by appendStoryToIndex
- [ ] list_stories returns story_type in each result object: Done — story_type returned in each list_stories result object via Story.StoryType
- [ ] list_stories accepts a story_type filter parameter: Done — story_type filter parameter added to list_stories tool
- [ ] get_story returns story_type in its result: Done — story_type returned in get_story result
- [ ] Build passes with no errors: Done — go build ./... passes with no errors

