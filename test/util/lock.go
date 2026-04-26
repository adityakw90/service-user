package testutil

import (
	"fmt"
	"os"
	"syscall"
)

// acquire lock to prevent multiple test processes from accessing the same resource
func AcquireLock(resource string) (*os.File, error) {
	path := fmt.Sprintf("/tmp/tests-svc-user-%s.lock", resource)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}

	return f, nil
}

// release lock on the resource
func ReleaseLock(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}
