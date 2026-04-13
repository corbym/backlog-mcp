package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const planScaffold = `# Plan

> **Status:** draft
>
> This plan is unfilled. Ask the user the questions in each section and complete it before creating stories.

## Overview

<!-- Ask: What is this project? What problem does it solve, and for whom? -->

## Goals

<!-- Ask: What does success look like? What are the key outcomes you're aiming for? -->

## Non-goals

<!-- Ask: What is explicitly out of scope? What should this project NOT do? -->

## Key constraints

<!-- Ask: Are there technical, timeline, resource, or other constraints to keep in mind? -->

## Open questions

<!-- Ask: What is still unresolved that will shape the work? -->
`

// runPlan creates a new plan file in the requirements directory.
// If name is given the file is plan-<name>.md, otherwise plan.md.
// If the target filename already exists, a numeric suffix is added (plan-002.md, etc.).
func runPlan(root, name string) error {
	path := nextPlanPath(root, name)
	if err := os.WriteFile(path, []byte(planScaffold), 0o644); err != nil {
		return err
	}
	fmt.Printf("  create %s\n", filepath.Base(path))
	fmt.Printf("\nOpen %s and work with your agent to fill it in.\n", filepath.Base(path))
	return nil
}

// nextPlanPath returns the first available plan filename in root.
// Sequence: plan.md → plan-002.md → plan-003.md (unnamed)
//
//	plan-<name>.md → plan-<name>-002.md → ... (named)
func nextPlanPath(root, name string) string {
	base := "plan"
	if name != "" {
		base = "plan-" + name
	}
	candidate := filepath.Join(root, base+".md")
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}
	for n := 2; ; n++ {
		candidate = filepath.Join(root, fmt.Sprintf("%s-%03d.md", base, n))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
