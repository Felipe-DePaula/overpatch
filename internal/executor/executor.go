// Package executor commits staged Overpatch file changes to disk.
package executor

import (
	"fmt"
	"io"
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
		info, err := existingFileInfo(path)
		if err != nil {
			return fmt.Errorf("applying %s: %w", change.Path, err)
		}
		if err := writeFileSafely(path, []byte(change.Staged), info.Mode().Perm()); err != nil {
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
		if err := writeFileSafely(path, []byte(change.Staged), 0o644); err != nil {
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
	_, err := existingFileInfo(path)
	return err
}

func existingFileInfo(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist")
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("target is not a file")
	}
	return info, nil
}

func writeFileSafely(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".overpatch-*")
	if err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	tempPath := temp.Name()

	n, err := temp.Write(data)
	if err != nil {
		if closeErr := temp.Close(); closeErr != nil {
			return errorWithTempCleanup(fmt.Errorf("writing temp file: %w; closing temp file after write failure: %v", err, closeErr), tempPath)
		}
		return errorWithTempCleanup(fmt.Errorf("writing temp file: %w", err), tempPath)
	}
	if n != len(data) {
		if closeErr := temp.Close(); closeErr != nil {
			return errorWithTempCleanup(fmt.Errorf("writing temp file: %w; closing temp file after short write: %v", io.ErrShortWrite, closeErr), tempPath)
		}
		return errorWithTempCleanup(fmt.Errorf("writing temp file: %w", io.ErrShortWrite), tempPath)
	}
	if err := temp.Chmod(perm); err != nil {
		if closeErr := temp.Close(); closeErr != nil {
			return errorWithTempCleanup(fmt.Errorf("setting permissions: %w; closing temp file after chmod failure: %v", err, closeErr), tempPath)
		}
		return errorWithTempCleanup(fmt.Errorf("setting permissions: %w", err), tempPath)
	}
	if err := temp.Close(); err != nil {
		return errorWithTempCleanup(fmt.Errorf("closing temp file: %w", err), tempPath)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return errorWithTempCleanup(fmt.Errorf("renaming temp file: %w", err), tempPath)
	}

	return nil
}

func errorWithTempCleanup(err error, tempPath string) error {
	if removeErr := os.Remove(tempPath); removeErr != nil && !os.IsNotExist(removeErr) {
		return fmt.Errorf("%w; removing temp file: %v", err, removeErr)
	}
	return err
}

func fullPath(root string, path string) string {
	return filepath.Join(root, filepath.FromSlash(path))
}
