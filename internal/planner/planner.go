// Package planner stages Overpatch operations in memory and renders previews.
package planner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Felipe-DePaula/overpatch/internal/schema"
)

const (
	StatusSuccess   = "success"
	StatusNoChanges = "no_changes"
)

// Result describes the outcome of a planning run.
type Result struct {
	Status       string
	Reason       string
	Operations   int
	FilesChanged int
	Diff         string
}

// OperationError identifies the operation that failed during staging.
type OperationError struct {
	OperationID string
	Message     string
}

func (err *OperationError) Error() string {
	return fmt.Sprintf("operation %s: %s", err.OperationID, err.Message)
}

type fileState struct {
	original       string
	staged         string
	existsOriginal bool
	existsStaged   bool
	loaded         bool
}

// Plan validates staging requirements, applies supported operations to
// in-memory file contents, and returns a textual preview.
func Plan(doc *schema.Document, root string) (*Result, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	if doc.Status == schema.StatusNoChanges {
		return &Result{
			Status:     StatusNoChanges,
			Reason:     doc.Reason,
			Operations: 0,
		}, nil
	}

	files := make(map[string]*fileState)

	for _, op := range doc.Operations {
		switch op.Action {
		case schema.ActionReplaceText:
			current, err := stagedContent(root, op.Path, files)
			if err != nil {
				return nil, operationError(op.ID, "reading file: %v", err)
			}

			actual := strings.Count(current, op.Find)
			if actual != op.ExpectedOccurrences {
				return nil, operationError(op.ID, "expected %d occurrence(s), found %d", op.ExpectedOccurrences, actual)
			}

			state := files[op.Path]
			state.staged = replaceText(current, op.Find, op.Replace, op.Occurrence)
			state.existsStaged = true
		case schema.ActionCreate:
			if err := stageCreate(root, op, files); err != nil {
				return nil, err
			}
		case schema.ActionDelete:
			if err := stageDelete(root, op, files); err != nil {
				return nil, err
			}
		default:
			return nil, operationError(op.ID, "action not supported by plan yet: %s", op.Action)
		}
	}

	changedPaths := changedFilePaths(files)
	diff := renderDiff(files, changedPaths)

	return &Result{
		Status:       StatusSuccess,
		Operations:   len(doc.Operations),
		FilesChanged: len(changedPaths),
		Diff:         diff,
	}, nil
}

func operationError(operationID string, format string, args ...any) *OperationError {
	return &OperationError{
		OperationID: operationID,
		Message:     fmt.Sprintf(format, args...),
	}
}

func stagedContent(root string, path string, files map[string]*fileState) (string, error) {
	state, err := loadFileState(root, path, files)
	if err != nil {
		return "", err
	}
	if !state.existsStaged {
		return "", os.ErrNotExist
	}
	return state.staged, nil
}

func stageCreate(root string, op schema.Operation, files map[string]*fileState) error {
	state, err := loadFileState(root, op.Path, files)
	if err != nil {
		return operationError(op.ID, "checking file: %v", err)
	}
	if state.existsStaged {
		return operationError(op.ID, "file already exists: %s", op.Path)
	}

	state.staged = *op.Content
	state.existsStaged = true
	return nil
}

func stageDelete(root string, op schema.Operation, files map[string]*fileState) error {
	state, err := loadFileState(root, op.Path, files)
	if err != nil {
		return operationError(op.ID, "checking file: %v", err)
	}
	if !state.existsStaged {
		return operationError(op.ID, "file does not exist: %s", op.Path)
	}

	info, err := os.Stat(fullPath(root, op.Path))
	if err != nil {
		if os.IsNotExist(err) {
			return operationError(op.ID, "file does not exist: %s", op.Path)
		}
		return operationError(op.ID, "checking file: %v", err)
	}
	if info.IsDir() {
		return operationError(op.ID, "delete target is not a file: %s", op.Path)
	}

	state.staged = ""
	state.existsStaged = false
	return nil
}

func loadFileState(root string, path string, files map[string]*fileState) (*fileState, error) {
	if state, ok := files[path]; ok {
		return state, nil
	}

	state := &fileState{loaded: true}
	files[path] = state

	info, err := os.Stat(fullPath(root, path))
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return nil, err
	}
	if info.IsDir() {
		state.existsOriginal = true
		state.existsStaged = true
		return state, nil
	}

	data, err := os.ReadFile(fullPath(root, path))
	if err != nil {
		return nil, err
	}

	state.original = string(data)
	state.staged = state.original
	state.existsOriginal = true
	state.existsStaged = true
	return state, nil
}

func fullPath(root string, path string) string {
	return filepath.Join(root, filepath.FromSlash(path))
}

func replaceText(content string, find string, replace string, occurrence string) string {
	if occurrence == "first" {
		return strings.Replace(content, find, replace, 1)
	}
	return strings.ReplaceAll(content, find, replace)
}

func changedFilePaths(files map[string]*fileState) []string {
	paths := make([]string, 0, len(files))
	for path, state := range files {
		if state.existsOriginal != state.existsStaged || state.original != state.staged {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func renderDiff(files map[string]*fileState, paths []string) string {
	var builder strings.Builder
	for i, path := range paths {
		if i > 0 {
			builder.WriteByte('\n')
		}
		state := files[path]

		builder.WriteString("--- ")
		if state.existsOriginal {
			builder.WriteString(path)
		} else {
			builder.WriteString("/dev/null")
		}
		builder.WriteByte('\n')
		builder.WriteString("+++ ")
		if state.existsStaged {
			builder.WriteString(path)
		} else {
			builder.WriteString("/dev/null")
		}
		builder.WriteByte('\n')
		builder.WriteString("@@\n")
		if state.existsOriginal {
			builder.WriteString(prefixLines("-", state.original))
		}
		if state.existsStaged {
			builder.WriteString(prefixLines("+", state.staged))
		}
	}
	return builder.String()
}

func prefixLines(prefix string, content string) string {
	if content == "" {
		return prefix + "\n"
	}

	lines := strings.SplitAfter(content, "\n")
	var builder strings.Builder
	for _, line := range lines {
		if line == "" {
			continue
		}
		builder.WriteString(prefix)
		builder.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}
