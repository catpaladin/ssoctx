// Package file contains needed functionality for config and files
package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
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
func AddLock(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	lock := lockFile{Time: time.Now()}
	lb, err := json.Marshal(lock)
	if err != nil {
		logger.Fatal().Msgf("Encountered error with marshal of temp lock file: %q", err)
	}

	if err := os.WriteFile(LockPath(), lb, 0o644); err != nil {
		logger.Fatal().Msgf("Encountered error writing temp lock file: %q", err)
	}
}

// RemoveLock removes a lock file
func RemoveLock(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	if err := os.Remove(LockPath()); err != nil {
		logger.Panic().Msgf("Encountered error removing temp lock file: %q", err)
	}
}

// LockStatus is used to lock a concurrent flow.
// e.g. Use to wrap authorization so ProcessClientInformation does not open a bunch of tabs
func LockStatus(ctx context.Context) bool {
	logger := zerolog.Ctx(ctx)

	var pathNotFound *os.PathError

	lb, err := os.ReadFile(LockPath())
	if err != nil {
		if errors.As(err, &pathNotFound) {
			return false
		}
		logger.Fatal().Msgf("Encountered error reading temp lock file %q", err)
	}

	lock := lockFile{}
	if err := json.Unmarshal(lb, &lock); err != nil {
		logger.Panic().Msgf("Encountered error while unmarshal of temp lock file: %q", err)
		return false
	}

	return time.Now().Before(lock.Time.Add(time.Minute))
}
