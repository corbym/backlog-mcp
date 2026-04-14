# STORY-019: Calibration report: compare pre-flight estimates vs actuals

**Type:** feature

## Goal

Build a report or tool output that surfaces estimate vs actual comparisons across all completed stories. Should show per-dimension accuracy and overall effort band accuracy. The output is used to identify which dimensions are most predictive and where the rubric needs adjustment.

## Acceptance criteria

- [ ] A get_estimation_report tool (or subcommand) exists that scans all completed stories for Estimate and Actuals sections
- [ ] Report output includes per-story: estimated band vs actual signals, and a flag if no estimate or actuals were recorded
- [ ] Aggregate summary shows: how many stories have both estimate+actuals, and mean absolute error per dimension where data exists
- [ ] Report runs without error when no stories have estimate/actuals data (returns empty summary)
- [ ] Output is human-readable text or structured JSON (agent-configurable)
