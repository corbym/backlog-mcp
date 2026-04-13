//go:build windows

package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const lockFileName = ".backlog-mcp.lock"
const lockPollInterval = 10 * time.Millisecond

// AcquireLock obtains an exclusive advisory lock on the requirements root
// directory, serialising all concurrent write operations.
//
// On Windows, flock(2) is not available. Instead, we attempt to create the
// lock file with O_CREATE|O_EXCL (which fails atomically if the file already
// exists), retrying every 10 ms until the lock is acquired or timeout elapses.
//
// The returned unlock function removes the lock file and must be called when
// the write is complete (typically via defer).
func AcquireLock(root string, timeout time.Duration) (unlock func(), err error) {
	path := filepath.Join(root, lockFileName)
	deadline := time.Now().Add(timeout)
	for {
		f, ferr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
		if ferr == nil {
			_ = f.Close()
			return func() {
				_ = os.Remove(path)
			}, nil
		}

		if !os.IsExist(ferr) {
			return nil, fmt.Errorf("acquiring backlog lock: %w", ferr)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for backlog lock after %s", timeout)
		}

		time.Sleep(lockPollInterval)
	}
}
