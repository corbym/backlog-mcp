# STORY-020: Automate pre-flight scoring on set_story_status in-progress

**Type:** feature

## Goal

Once the rubric is calibrated from real data, automate the pre-flight scoring step. When an agent calls set_story_status in-progress, the server should analyse the story content (AC count, description length, referenced files) and return an automated effort estimate, removing the need for the agent to manually score each dimension.

## Acceptance criteria

- [ ] set_story_status in-progress response includes an auto-generated effort estimate based on story content analysis
- [ ] Analysis inspects: AC count, description word count, number of referenced files/symbols, story_type
- [ ] Returned estimate includes a band (S/M/L/XL) and a brief rationale explaining the score
- [ ] Auto-score is clearly labelled as automated (not agent-provided) in the story file if persisted
- [ ] Accuracy baseline: auto-score matches agent self-estimate within one band on ≥70% of the calibration dataset from STORY-019

## Notes

<!-- backlog-mcp: 2026-05-03T21:53:41Z -->
Deferred. Automated pre-flight scoring against a calibrated rubric was the end-state of the estimation epic. In the reframed epic, STORY-018 covers automated readiness flag detection on set_story_status in-progress — this story's goal is absorbed by that work. Revisit only if STORY-018 turns out to need a separate automation step.
