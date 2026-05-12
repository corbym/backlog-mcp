# STORY-023: Create Homebrew tap formula and release automation

## Description

Create a Homebrew tap formula for `backlog-mcp` so users can install via `brew install corbym/backlog-mcp/backlog-mcp`. Wire up GitHub Actions to automatically update the tap when a new release is published.

## Acceptance Criteria

- [x] `.goreleaser.yaml` produces versioned archives (e.g. `backlog-mcp_1.0.0_darwin_amd64.tar.gz`) and a `checksums.txt`
- [x] `main.go` supports a `--version` flag that prints the injected version
- [x] `.github/workflows/update-homebrew.yml` exists in `corbym/backlog-mcp` and updates `corbym/homebrew-backlog-mcp/Formula/backlog-mcp.rb` on dispatch
- [x] `.github/workflows/release.yml` triggers `update-homebrew.yml` after a successful goreleaser run
- [x] `Formula/backlog-mcp.rb` template is documented so it can be seeded into `corbym/homebrew-backlog-mcp`

## Notes

<!-- backlog-mcp: 2026-05-12T08:32:04Z -->
Story created for PR on branch copilot/create-homebrew-tap-formula. Branch was created before story per task requirement; story ID STORY-023 assigned retroactively.

<!-- backlog-mcp: 2026-05-12T08:59:58Z -->
PR #9: feat(STORY-023): Add Homebrew tap formula and release automation
