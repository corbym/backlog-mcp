# STORY-018: Agent pre-flight self-estimate recorded on story start

**Type:** feature

## Goal

Before an agent sets a story to in-progress, it should score the story against each rubric dimension and record that estimate in the story file. The estimate becomes a pre-flight record that can later be compared against actuals. Format: structured note with per-dimension scores and a total estimated effort band (S/M/L/XL).

## Acceptance criteria

- [ ] When set_story_status in-progress is called, the tool response includes a prompt reminding the agent to record a pre-flight estimate
- [ ] A new add_story_estimate tool (or extended add_story_note) accepts per-dimension scores and an overall S/M/L/XL band
- [ ] The estimate is stored in the story file under a `## Estimate` section with a timestamp
- [ ] Calling the estimate tool is not mandatory — the in-progress transition is not blocked if skipped
- [ ] Estimate data is distinct from notes and actuals so it can be parsed independently
