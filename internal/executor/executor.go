// Package executor commits staged Overpatch file changes to disk.
package executor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Felipe-DePaula/overpatch/internal/planner"
)

// Apply writes a prepared StageResult to the filesystem under root.
func Apply(stage *planner.StageResult, root string) error {
	if stage == nil {
		return fmt.Errorf("stage is nil")
	}

	for _, change := range stage.Changes {
		if err := applyFileChange(change, root); err != nil {
			return err
		}
	}

	return nil
}

func applyFileChange(change planner.FileChange, root string) error {
	path := fullPath(root, change.Path)

	switch change.Kind {
	case planner.FileChangeModified:
		if err := ensureExistingFile(path); err != nil {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}
		if err := os.WriteFile(path, []byte(change.Staged), 0o600); err != nil {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}
	case planner.FileChangeCreated:
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("applying %s: file already exists", change.Path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}
		if err := os.WriteFile(path, []byte(change.Staged), 0o600); err != nil {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}
	case planner.FileChangeDeleted:
		if err := ensureExistingFile(path); err != nil {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}
	default:
		return fmt.Errorf("applying %s: unsupported file change kind: %s", change.Path, change.Kind)
	}

	return nil
}

func ensureExistingFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist")
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("target is not a file")
	}
	return nil
}

func fullPath(root string, path string) string {
	return filepath.Join(root, filepath.FromSlash(path))
}
