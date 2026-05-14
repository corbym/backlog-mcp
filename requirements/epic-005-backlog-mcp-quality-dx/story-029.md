# STORY-029: Fix backlog.go regex to handle format deviations gracefully

**Type:** bug

## Goal

The backlog entry regex assumes exact format `**STORY-NNN** — description` (with an em-dash). Any deviation causes entries to be passed through unmodified with no error or warning. Audit the regex against real backlog entries, tighten or broaden the match as needed, and add a warning/error when an expected entry cannot be found.

## Acceptance criteria

- [ ] Define acceptance criteria
