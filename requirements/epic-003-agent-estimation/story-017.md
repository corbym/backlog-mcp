# STORY-017: Capture actuals on story completion

**Type:** feature

## Goal

When an agent completes a story, capture measurable signals and store them as structured data alongside the completion note. Signals to capture: number of tool calls, files read, files edited, number of test files changed, and whether the AC had to be revised mid-story. This raw data is the foundation for calibrating the rubric.

## Acceptance criteria

- [ ] When complete_story is called, the completion note includes a structured `## Actuals` section
- [ ] Actuals captured: tool calls count, files read count, files edited count, test files changed count, AC revised (bool)
- [ ] Actuals are stored in the story file alongside the completion summary note
- [ ] If actuals data is not provided by the agent, the section is omitted (not an error) — capture is best-effort
- [ ] Existing complete_story tests still pass
