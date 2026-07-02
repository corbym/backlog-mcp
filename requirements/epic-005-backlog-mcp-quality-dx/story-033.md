# STORY-033: Fix bulk_update_acceptance_criteria combined ID:text key lookup failure

**Type:** bug

## Goal

bulk_update_acceptance_criteria rejects criterion keys of the form `AC-STORY-NNN-xxxxxxxx: full criterion text` (ID + colon + description), returning 'criterion/criteria not found' for all of them. The lookup in PatchAcceptanceCriteria (parser/story.go) checks acByID[key] — but the key is the full combined string, not just the ID. It also checks acByText[normalizeACKey(key)] — but the stored text is just the description without the ID prefix. Neither matches. This is the natural key format an agent composes from a get_story response. Fix: call parseACText(key) before the lookup to extract the bare ID, then look that up in acByID.

## Acceptance criteria

- [ ] AC-STORY-033-cfe040f9: bulk_update_acceptance_criteria called with a key of the form `AC-STORY-NNN-xxxxxxxx: full criterion text` successfully matches and updates the criterion
- [ ] AC-STORY-033-73b88301: bulk_update_acceptance_criteria called with a bare ID `AC-STORY-NNN-xxxxxxxx` continues to work as before
- [ ] AC-STORY-033-5509c0cf: bulk_update_acceptance_criteria called with plain criterion text (no ID prefix) continues to work as before
- [ ] AC-STORY-033-e1c11d34: A failing test is written first reproducing the 'criterion not found' error when a combined ID:text key is passed
- [ ] AC-STORY-033-1683d659: The fix is in PatchAcceptanceCriteria in parser/story.go — parseACText(key) is called before the acByID lookup
- [ ] AC-STORY-033-f8c8bcc9: All existing tests pass
