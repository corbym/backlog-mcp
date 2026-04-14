# STORY-011: Document all 9 MCP tools and fix set_acceptance_criteria schema bug

## Goal

Ensure all tools have high-quality descriptions and parameter documentation for glama.ai scoring. Fix the missing criteria array parameter in the set_acceptance_criteria tool schema. Update CLAUDE.md to document the 4 tools added since it was last written.

## Acceptance criteria

- [x] All 9 tools have descriptions that include what they return
- [x] All required parameters are marked required in the schema
- [x] set_acceptance_criteria criteria array parameter is present in tool schema
- [x] All optional parameters have clear descriptions explaining their purpose
- [x] CLAUDE.md documents all 9 tools (not just the original 5)
- [x] CLAUDE.md design decisions section reflects current implementation (file locking)
- [x] Build passes with no errors

## Notes

<!-- backlog-mcp: 2026-04-14T09:18:19Z -->
Updated all 9 tool descriptions in tools.go to include return value shapes and clearer parameter descriptions. Fixed a bug where the `criteria` array parameter was entirely absent from the `set_acceptance_criteria` tool schema (the handler used it but it was never declared, so MCP clients had no way to know it existed). Updated CLAUDE.md: corrected "five tools" to "nine tools", added full documentation for create_epic, create_story, set_acceptance_criteria, and complete_story, and corrected the design decisions section to reflect that file locking is already implemented. Build verified clean.

Incomplete criteria:
- [ ] All 9 tools have descriptions that include what they return: Done — all 9 tool descriptions updated in tools.go to include return value shape
- [ ] All required parameters are marked required in the schema: Done — all required parameters verified with mcp.Required(); set_acceptance_criteria criteria param now marked Required()
- [ ] set_acceptance_criteria criteria array parameter is present in tool schema: Done — mcp.WithArray("criteria", ..., mcp.Required()) added to set_acceptance_criteria tool registration
- [ ] All optional parameters have clear descriptions explaining their purpose: Done — all optional parameter descriptions updated to explain purpose and valid values
- [ ] CLAUDE.md documents all 9 tools (not just the original 5): Done — CLAUDE.md updated with full documentation for all 9 tools including the 4 previously missing
- [ ] CLAUDE.md design decisions section reflects current implementation (file locking): Done — CLAUDE.md design decisions updated: 'No locking in PoC' replaced with accurate description of AcquireLock implementation
- [ ] Build passes with no errors: Done — go build ./... passes with no errors

