# STORY-035: get_story and bulk_update_acceptance_criteria return full unbounded file content

**Type:** chore

## Goal

`get_story` always returns the entire raw markdown of a story file, with no summary or metadata-only mode. Story files only grow over time — every `add_story_note` and `complete_story` call appends a timestamped note under `## Notes`, and nothing ever trims it — so a heavily-annotated, long-lived story can be many KB, and every `get_story` call pays for its full history even when the caller only needs current status or acceptance criteria. `bulk_update_acceptance_criteria` has the same shape: it echoes the entire post-update file content back in its response (`tools.go`, the `"content": content` field) even though callers generally only need to know which criteria changed and which didn't match.

This was flagged reviewing agent token usage against a real dogfood session (recordari repo, 2026-07-09): repeated `get_story` calls on a story with multiple appended completion/verification notes, plus a `bulk_update_acceptance_criteria` response, made up a disproportionate share of that session's backlog-mcp token spend.

## Acceptance criteria

- [x] AC-STORY-035-7d5dbddd: bulk_update_acceptance_criteria's response no longer includes the `content` field — returns `{story_id, path, criteria_updated, errors}`
- [x] AC-STORY-035-800e66ab: bulk_update_acceptance_criteria's tool description is updated to match the new response shape
- [x] AC-STORY-035-85352f80: get_story accepts an optional `include_notes` boolean parameter, default `true`
- [x] AC-STORY-035-1f334cda: get_story called with `include_notes=false` on a story containing a `## Notes` section returns `content` truncated to everything before that heading, with trailing whitespace trimmed
- [x] AC-STORY-035-9412948d: get_story called with `include_notes=false` on a story with no `## Notes` section returns the full content unchanged
- [x] AC-STORY-035-039a8553: get_story called with `include_notes` omitted or `true` is byte-for-byte unchanged from current behaviour (full content, including notes)
- [x] AC-STORY-035-909ed238: get_story's tool description documents the `include_notes` parameter and its default
- [x] AC-STORY-035-03e83f37: Failing tests are written first for both changes, per the domain's TDD-mandatory standing rule
- [x] AC-STORY-035-a853612c: All existing tests pass

## Notes

Design decided 2026-07-09: get_story gets an optional `include_notes` param (default `true`, fully backward-compatible) rather than a new `get_story_summary` tool — keeps the tool surface at nine per CLAUDE.md rather than growing it. Split point for the notes truncation is the first `## Notes` heading, consistent with how `AppendNote` in `parser/story.go` locates the section for writes. `bulk_update_acceptance_criteria` drops the `content` echo outright (option b from the original triage) — `criteria_updated`/`errors` is already sufficient signal; callers who want to see the result call `get_story` explicitly.

<!-- backlog-mcp: 2026-07-09T00:00:00Z -->
Implemented in `tools.go`. `get_story` gained an optional `include_notes` boolean param (`req.GetBool("include_notes", true)`); when false, content is sliced at `strings.Index(content, "## Notes")` and right-trimmed. `bulk_update_acceptance_criteria` no longer builds or returns a `content` field — the now-unused `parser.ReadStory` call after `PatchAcceptanceCriteria` was removed entirely. Both tool descriptions updated. TDD: failing tests written first (`tools_test.go`: `TestGetStory_IncludeNotesFalse_TruncatesAtNotesHeading`, `TestGetStory_IncludeNotesFalse_NoNotesSection_ReturnsFullContent`, `TestGetStory_IncludeNotesOmitted_ReturnsFullContent`, and the flipped assertion in `TestBulkUpdateAC_ChecksNamedCriteria`), confirmed failing against the old code, then the implementation made them pass. Full suite green (`go test ./...`).
