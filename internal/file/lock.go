// Package file contains needed functionality for config and files
package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

type lockFile struct {
	Time time.Time `json:"Time"`
}

// LockPath returns the full path to expected lock file
func LockPath() string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("%s\\aws-sso-util.lock", os.TempDir())
	default:
		return fmt.Sprintf("%s/aws-sso-util.lock", os.TempDir())
	}
}

// AddLock creates a lock file
func AddLock() {
	lock := lockFile{Time: time.Now()}
	lb, err := json.Marshal(lock)
	if err != nil {
		log.Fatalf("Encountered error with marshal of temp lock file: %q", err)
	}

	if err := os.WriteFile(LockPath(), lb, 0o644); err != nil {
		log.Fatalf("Encountered error writing temp lock file: %q", err)
	}
}

// RemoveLock removes a lock file
func RemoveLock() {
	if err := os.Remove(LockPath()); err != nil {
		log.Panicf("Encountered error removing temp lock file: %q", err)
	}
}

// LockStatus is used to lock a concurrent flow.
// e.g. Use to wrap authorization so ProcessClientInformation does not open a bunch of tabs
func LockStatus() bool {
	var pathNotFound *os.PathError

	lb, err := os.ReadFile(LockPath())
	if err != nil {
		if errors.As(err, &pathNotFound) {
			return false
		}
		log.Fatalf("Encountered error reading temp lock file %q", err)
	}

	lock := lockFile{}
	if err := json.Unmarshal(lb, &lock); err != nil {
		log.Panicf("Encountered error while unmarshal of temp lock file: %q", err)
		return false
	}

	return time.Now().Before(lock.Time.Add(time.Minute))
}
