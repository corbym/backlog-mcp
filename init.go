package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const scaffoldIndex = `# Requirements Index

## EPIC-001: Example Epic — ` + "`draft`" + `

| Story | Title | Status |
|-------|-------|--------|
| [STORY-001](./epic-001-example/story-001.md) | Example story | draft |
`

const scaffoldBacklog = `# Backlog

Stories are listed in priority order (head = highest priority).
When a story reaches ` + "`done`" + `, remove it from the backlog and renumber.

1. **STORY-001** — Example story
`

const scaffoldStory = `# STORY-001: Example story

## Goal

Describe what this story should accomplish.

## Acceptance criteria

- [ ] Criterion one
- [ ] Criterion two
`

const agentsSection = `
## Backlog (backlog-mcp)

Project stories and epics live in ` + "`requirements/`" + `. Use the backlog-mcp MCP tools to interact with them — do not edit the index or backlog files by hand.

**Tools**

- ` + "`list_stories`" + ` — list stories, optionally filtered by ` + "`epic_id`" + ` or ` + "`status`" + `
- ` + "`get_story`" + ` — read the full markdown content and metadata for a story
- ` + "`set_story_status`" + ` — update a story's status (` + "`draft`" + `, ` + "`in-progress`" + `, ` + "`done`" + `, ` + "`blocked`" + `)
- ` + "`add_story_note`" + ` — append a timestamped note to a story file
- ` + "`get_index_summary`" + ` — high-level epic/story counts by status

**Conventions**

- Story IDs: ` + "`STORY-NNN`" + ` (e.g. ` + "`STORY-047`" + `)
- Epic IDs: ` + "`EPIC-NNN`" + ` (e.g. ` + "`EPIC-003`" + `)
- Always call ` + "`set_story_status`" + ` when you start or finish work on a story
- Use ` + "`add_story_note`" + ` to record decisions, blockers, or progress — never edit story files directly

**Plans**

Plans live in ` + "`requirements/plan*.md`" + `. They describe overall project goals and direction.

- Read all plan files before creating or prioritising stories
- When you find a plan marked ` + "`Status: draft`" + `, ask the user the questions in each section and fill it in before proceeding
- Use plans to decide what stories to create and how to order the backlog
- Plans never reference specific stories — keep that relationship one-way (stories may reference plan sections, not the other way around)
- Never edit a plan to add story references or implementation detail — plans are for goals and intent, not execution tracking
`

// runInit creates a new backlog directory at the given path and populates it
// with the required files so the MCP server can start immediately.
func runInit(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %q: %w", dir, err)
	}

	files := []struct {
		rel     string
		content string
	}{
		{"requirements-index.md", scaffoldIndex},
		{"backlog.md", scaffoldBacklog},
		{"epic-001-example/story-001.md", scaffoldStory},
	}

	for _, f := range files {
		path := filepath.Join(dir, filepath.FromSlash(f.rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("creating directory for %q: %w", f.rel, err)
		}
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  skip  %s (already exists)\n", f.rel)
			continue
		}
		if err := os.WriteFile(path, []byte(f.content), 0o644); err != nil {
			return fmt.Errorf("writing %q: %w", f.rel, err)
		}
		fmt.Printf("  create %s\n", f.rel)
	}

	if err := appendAgentsSection(filepath.Dir(dir)); err != nil {
		return fmt.Errorf("writing AGENTS.md: %w", err)
	}

	fmt.Printf("\nBacklog initialised at %s\n", dir)
	fmt.Printf("Run ./backlog-mcp from the project root to start the server.\n")
	return nil
}

// appendAgentsSection creates or appends the backlog instructions to AGENTS.md
// in the given directory (expected to be the project root).
func appendAgentsSection(projectRoot string) error {
	path := filepath.Join(projectRoot, "AGENTS.md")

	// If the file exists, only append if our section isn't already there.
	if existing, err := os.ReadFile(path); err == nil {
		if strings.Contains(string(existing), "## Backlog (backlog-mcp)") {
			fmt.Printf("  skip  AGENTS.md (backlog section already present)\n")
			return nil
		}
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(agentsSection)
		fmt.Printf("  update AGENTS.md\n")
		return err
	}

	if err := os.WriteFile(path, []byte("# Agent Instructions\n"+agentsSection), 0o644); err != nil {
		return err
	}
	fmt.Printf("  create AGENTS.md\n")
	return nil
}
