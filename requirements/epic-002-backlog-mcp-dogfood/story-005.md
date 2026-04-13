# STORY-005: Fix hardcoded http:// base URL in server.go

## Goal

`server.go` passes a hardcoded `http://` base URL to the MCP server constructor. This will silently serve insecure URLs if the server is exposed over HTTPS. Fix it to derive the scheme from config or environment.

## Acceptance criteria

- [ ] Base URL scheme is configurable via environment variable (e.g. `BACKLOG_BASE_URL` or `BACKLOG_SCHEME`)
- [ ] Defaults to `http` for local use, not hardcoded
- [ ] HTTP/SSE mode works correctly after the change
- [ ] README documents the new variable