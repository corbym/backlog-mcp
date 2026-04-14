# STORY-016: Spike: define estimation dimensions and scoring rubric

**Type:** spike

## Goal

Research and document the dimensions that drive agent implementation effort. Produce a draft rubric with scoring guidance for each dimension: context surface (files to read), edit surface (files to change), AC ambiguity, pattern familiarity, and test complexity. The rubric should be simple enough for an agent to apply in under a minute before starting a story.

## Acceptance criteria

- [ ] A markdown document exists defining each estimation dimension with a clear description and 1–5 scoring guidance
- [ ] Dimensions covered: context surface, edit surface, AC ambiguity, pattern familiarity, test complexity
- [ ] Each score level (1–5) has at least one concrete example drawn from this codebase
- [ ] A worked example scores an existing completed story (e.g. STORY-007 or STORY-009) against the full rubric
- [ ] The rubric document is committed to the requirements directory and linked from epic-003.md
