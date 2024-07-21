package file

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type MockFileSystem struct {
	WriteFileFunc func(name string, data []byte, perm os.FileMode) error
	ReadFileFunc  func(name string) ([]byte, error)
	RemoveFunc    func(name string) error
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return m.WriteFileFunc(name, data, perm)
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	return m.ReadFileFunc(name)
}

func (m *MockFileSystem) Remove(name string) error {
	return m.RemoveFunc(name)
}

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestAddLock(t *testing.T) {
	mockFS := &MockFileSystem{
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error {
			return nil
		},
	}

	SetFileSystem(mockFS)

	ctx := log.Logger.With().Logger().WithContext(context.Background())

	AddLock(ctx)
}

func TestRemoveLock(t *testing.T) {
	mockFS := &MockFileSystem{
		RemoveFunc: func(name string) error {
			return nil
		},
	}

	SetFileSystem(mockFS)

	ctx := log.Logger.With().Logger().WithContext(context.Background())

	RemoveLock(ctx)
}

func TestIsLocked(t *testing.T) {
	lockTime := time.Now()
	lockData, _ := json.Marshal(lockFile{Time: lockTime})

	tests := []struct {
		name       string
		readFile   func(name string) ([]byte, error)
		wantLocked bool
	}{
		{
			name: "not locked - file not found",
			readFile: func(name string) ([]byte, error) {
				return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
			},
			wantLocked: false,
		},
		{
			name: "locked - expired lock",
			readFile: func(name string) ([]byte, error) {
				return lockData, nil
			},
			wantLocked: true,
		},
		{
			name: "locked - valid lock",
			readFile: func(name string) ([]byte, error) {
				lockData, _ := json.Marshal(lockFile{Time: time.Now().Add(time.Minute)})
				return lockData, nil
			},
			wantLocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := &MockFileSystem{
				ReadFileFunc: tt.readFile,
			}

			SetFileSystem(mockFS)

			ctx := log.Logger.With().Logger().WithContext(context.Background())

			if got := IsLocked(ctx); got != tt.wantLocked {
				t.Errorf("IsLocked() = %v, want %v", got, tt.wantLocked)
			}
		})
	}
}
