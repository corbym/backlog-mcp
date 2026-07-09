# STORY-034: acIDRe requires an 8-hex-char suffix, so plain sequential AC IDs never resolve via acByID

**Type:** bug

## Goal

`acIDRe` in `parser/story.go` is `^(AC-[A-Z]+-\d+-[0-9a-fA-F]{8}): (.*)$` — it only recognises AC IDs with an 8-character hex suffix (e.g. `AC-STORY-042-a3f9b2c1`). Stories whose acceptance criteria were hand-authored with plain sequential IDs (e.g. `AC-STORY-159-1`, `AC-STORY-159-2`, ...) instead of the tool-generated hex-suffixed format never match this regex, so `parseACText` returns `("", fullText)` for every one of them. `PatchAcceptanceCriteria` in `bulk_update_acceptance_criteria` then builds an empty `acByID` map for that story — a bare-ID key like `"AC-STORY-159-1"` misses both `acByID` (never populated) and `acByText` (keyed on the full untrimmed criterion text, not the ID), and the call fails with "criterion/criteria not found" for every criterion in the story, even though the same bare-ID call works fine for stories whose ACs do carry hex suffixes.

This is distinct from STORY-033 (combined `"ID: text"` key lookup) — this bug affects *bare* ID lookups when the story's own AC IDs were never hex-suffixed in the first place, most commonly for stories authored by an external tool/agent that writes AC lines directly to the markdown rather than going through `set_acceptance_criteria`.

Observed 2026-07-09 filing STORY-159 in the recordari repo's backlog-mcp instance: all 12 of its plain-numbered ACs (`AC-STORY-159-1` .. `AC-STORY-159-12`) failed `bulk_update_acceptance_criteria`, requiring a fallback to 12 individual `check_acceptance_criterion(criterion_index=N)` calls.

## Acceptance criteria

- [x] AC-STORY-034-4b2e9a17: parseACText recognises a plain sequential AC ID suffix (e.g. `AC-STORY-159-1`, `AC-STORY-159-12`) and returns the full ID and stripped text, not `("", fullText)`
- [x] AC-STORY-034-6f81c3d0: parseACText continues to recognise existing 8-hex-char-suffixed IDs (e.g. `AC-STORY-042-a3f9b2c1`) — regression guard
- [x] AC-STORY-034-2c94e6b5: bulk_update_acceptance_criteria successfully matches and updates a criterion by its bare plain-sequential ID
- [x] AC-STORY-034-d70a1f89: check_acceptance_criterion successfully matches a criterion by its bare plain-sequential ID (ID-based lookup already shares parseACText, so this should pass as a side effect of the same fix — verified explicitly here)
- [x] AC-STORY-034-fe3b58c2: A failing test is written first reproducing the "criterion not found" error when a plain-sequential AC ID is passed to bulk_update_acceptance_criteria
- [x] AC-STORY-034-91d4a7e6: All existing tests pass

## Notes

Design decided 2026-07-09: fix is option (a) — relax `acIDRe` in `parser/story.go` from `^(AC-[A-Z]+-\d+-[0-9a-fA-F]{8}): (.*)$` to `^(AC-[A-Z]+-\d+-[0-9a-fA-F]+): (.*)$` (exact-8 → one-or-more hex chars). Decimal digits are a subset of hex digits, so this one-line change accepts both the tool-generated 8-hex-char suffix and hand-authored plain sequential suffixes (`-1`, `-2`, ... `-12`) without a separate fallback path. Because `CheckAcceptanceCriterion`'s ID-based matching (`matchByID` in `parser/story.go`) already calls through the same `parseACText`/`acByID` machinery, `check_acceptance_criterion` gets ID-based lookup for plain sequential IDs for free — no separate implementation needed, just a regression test to confirm it.

<!-- backlog-mcp: 2026-07-09T00:00:00Z -->
Fixed by relaxing `acIDRe`'s hex suffix from exactly 8 chars to one-or-more (`[0-9a-fA-F]{8}` → `[0-9a-fA-F]+`) in `parser/story.go`. TDD: failing tests written first (`parser/story_test.go`: `TestParseACText_PlainSequentialID`, `TestParseACText_PlainSequentialID_MultiDigit`; `tools_test.go`: `TestBulkUpdateAC_MatchByPlainSequentialID`, `TestCheckAcceptanceCriterion_MatchByPlainSequentialID`), confirmed failing against the old regex, then the one-line fix made them pass. `check_acceptance_criterion` ID-based matching came free from the same `parseACText` change, confirmed by regression test. Full suite green (`go test ./...`). Tool descriptions for `bulk_update_acceptance_criteria` and `check_acceptance_criterion` were already ID-format-agnostic, no doc changes needed.
