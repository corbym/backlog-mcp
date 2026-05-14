# STORY-024: Add reorder_backlog MCP tool

**Type:** feature

## Goal

Add a new `reorder_backlog` MCP tool that accepts an ordered list of story IDs and rewrites `backlog.md` to match. Must go through the same atomic write + file lock path as all other mutating tools. Agents currently reorder the backlog via direct file edits, bypassing safety guarantees.

## Acceptance criteria

- [ ] Define acceptance criteria
