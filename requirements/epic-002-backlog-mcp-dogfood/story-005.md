# STORY-005: Fix hardcoded http:// base URL in server.go

## Goal

`server.go` passes a hardcoded `http://` base URL to the MCP server constructor. This will silently serve insecure URLs if the server is exposed over HTTPS. Fix it to derive the scheme from config or environment.

## Acceptance criteria

- [x] HTTP/SSE mode removed entirely — base URL hardcoding is a non-issue

## Notes

<!-- backlog-mcp: 2026-04-14T10:15:15Z -->
Dropped. HTTP/SSE mode was removed from the codebase entirely (runHTTP, WithBaseURL, and the BACKLOG_TRANSPORT switch are all gone). The hardcoded http:// issue is moot. Original ACs replaced to reflect the actual resolution.
