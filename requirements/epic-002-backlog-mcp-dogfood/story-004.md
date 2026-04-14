# STORY-004: Add file locking for concurrent agent writes

## Goal

The server currently assumes a single writer. If two agents call `set_story_status` or `add_story_note` simultaneously, writes may interleave and corrupt the index or backlog. Add a file lock to serialise mutations.

## Acceptance criteria

- [x] A file-based lock (e.g. `flock` or `.lock` sentinel) is acquired before any write operation
- [x] Lock is released after the atomic rename completes
- [x] Concurrent writes serialise correctly under test (e.g. two goroutines hammering `set_story_status`)
- [x] Lock timeout returns a clear error rather than hanging indefinitely
## Notes

<!-- backlog-mcp: 2026-04-13T14:49:53Z -->
## Technical approach

- New `parser/lock.go`: `AcquireLock(root string, timeout time.Duration) (unlock func(), error)`
- Lock target: `.backlog-mcp.lock` file in the requirements root (created on first use, never deleted)
- Mechanism: `syscall.Flock(LOCK_EX | LOCK_NB)` in a 10ms poll loop until acquired or timeout
- On timeout: returns clear error `"timed out waiting for backlog lock after Xs"`
- Callers: `set_story_status`, `add_story_note`, `create_epic`, `create_story` in `tools.go` — acquire at handler entry, `defer unlock()`
- Read-only tools (`list_stories`, `get_story`, `get_index_summary`) not locked
- No new dependencies — `syscall` is stdlib, works on macOS/Linux
- Default timeout: 5 seconds
- New test: `TestConcurrentSetStoryStatus_Serialises` — 10 goroutines hammer `set_story_status` concurrently, assert index file is uncorrupted afterwards
