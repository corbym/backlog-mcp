//go:build !windows

package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const lockFileName = ".backlog-mcp.lock"
const lockPollInterval = 10 * time.Millisecond

// AcquireLock obtains an exclusive advisory lock on the requirements root
// directory, serialising all concurrent write operations.
//
// It opens (or creates) a sentinel file and calls syscall.Flock with
// LOCK_EX|LOCK_NB, retrying every 10 ms until the lock is acquired or
// timeout elapses.
//
// The returned unlock function releases the lock and must be called when the
// write is complete (typically via defer).
func AcquireLock(root string, timeout time.Duration) (unlock func(), err error) {
	path := filepath.Join(root, lockFileName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening lock file: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for {
		ferr := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if ferr == nil {
			// Lock acquired.
			return func() {
				_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				_ = f.Close()
			}, nil
		}

		if ferr != syscall.EWOULDBLOCK {
			_ = f.Close()
			return nil, fmt.Errorf("acquiring backlog lock: %w", ferr)
		}

		if time.Now().After(deadline) {
			_ = f.Close()
			return nil, fmt.Errorf("timed out waiting for backlog lock after %s", timeout)
		}

		time.Sleep(lockPollInterval)
	}
}
