# EPIC-003: Agent Estimation

## Goal

Define how to estimate the effort and duration of AI agent work on backlog stories. The goal is to move from "no idea" to a calibrated rubric agents can apply pre-flight, backed by measured actuals. Key dimensions to explore: context surface (how much the agent must read), edit surface (files changed), ambiguity of ACs, pattern familiarity, and test complexity. Approach: instrument actuals first, derive rubric from data, then automate pre-flight scoring.

## Stories

- [ ] [STORY-016](story-016.md) — Spike: define estimation dimensions and scoring rubric
- [ ] [STORY-017](story-017.md) — Capture actuals on story completion
- [ ] [STORY-018](story-018.md) — Agent pre-flight self-estimate recorded on story start
- [ ] [STORY-019](story-019.md) — Calibration report: compare pre-flight estimates vs actuals
- [ ] [STORY-020](story-020.md) — Automate pre-flight scoring on set_story_status in-progress
