# STORY-025: Fix em-dash encoding artefacts in acceptance criteria text

**Type:** bug

## Goal

Em-dashes in acceptance criteria text are stored with encoding artefacts, causing `bulk_update_acceptance_criteria` text matching to silently fail with `criteria_errors`. Fix in the file write layer so em-dashes (and other Unicode punctuation) round-trip cleanly. Forces agents into index-based workarounds and will recur since em-dashes are natural in criterion descriptions.

## Acceptance criteria

- [ ] Define acceptance criteria
