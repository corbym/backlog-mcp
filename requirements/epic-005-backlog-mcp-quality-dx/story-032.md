# STORY-032: Fix list_stories Unicode encoding artefacts in story titles

**Type:** bug

## Goal

list_stories returns story titles containing em-dashes (—) and arrows (→) as UTF-8 replacement characters (� / U+FFFD), e.g. 'Walking skeleton ��� remember + recall'. Likely a double-encoding issue in the file read path (UTF-8 bytes read as Latin-1 or Windows-1252 then re-encoded). Makes list_stories output unreliable for display or downstream parsing and suggests a broader encoding inconsistency in how story files are read on Windows.

## Acceptance criteria

- [ ] AC-STORY-032-5a966074: list_stories returns story titles containing em-dashes (—) without replacement-character artefacts
- [ ] AC-STORY-032-855a631a: list_stories returns story titles containing arrows (→) without replacement-character artefacts
- [ ] AC-STORY-032-e6474f62: A failing test is written first using a story file whose title contains at least one multi-byte Unicode character
- [ ] AC-STORY-032-80431a72: The fix is applied in the file read path so all story file reads are consistently UTF-8
- [ ] AC-STORY-032-46939e59: All existing tests pass
