# STORY-016: Spike: define story readiness criteria and risk flags

**Type:** spike

## Goal

Research and document what makes a story ready to implement. Identify the conditions whose absence predicts a false start, a mid-story AC revision, or a scope explosion. Produce a short written definition of "story readiness" with a checklist of concrete, detectable flags an agent can evaluate in under a minute before starting work. The checklist is the foundation for the automated readiness check in STORY-018.

## Acceptance criteria

- [ ] A markdown document exists defining "story readiness" with a checklist of named risk flags
- [ ] Flags covered include at minimum: AC missing or placeholder, AC are vague or untestable, story has unresolved open questions (e.g. "?" in notes), test expectations are not explicit, an obvious dependency is not yet done
- [ ] Each flag has: a name, how to detect it (heuristic or exact check), severity (advisory vs blocking), and a concrete example from this codebase showing a story with and without the flag
- [ ] A worked example runs the checklist against one completed story from EPIC-002 and shows which flags fired and which cleared
- [ ] The document is committed to requirements/epic-003-agent-estimation/ and linked from epic-003.md
