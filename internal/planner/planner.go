// Package planner stages Overpatch operations in memory and renders previews.
package planner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Felipe-DePaula/overpatch/internal/diff"
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
		case schema.ActionReplaceLines:
			current, err := stagedContent(root, op.Path, files)
			if err != nil {
				return nil, operationError(op.ID, "reading file: %v", err)
			}

			next, actual := replaceLineBlock(current, op.FindLines, op.ReplaceLines)
			if actual != op.ExpectedOccurrences {
				return nil, operationError(op.ID, "expected %d occurrence(s), found %d", op.ExpectedOccurrences, actual)
			}

			state := files[op.Path]
			state.staged = next
			state.existsStaged = true
		case schema.ActionInsertBeforeLines:
			current, err := stagedContent(root, op.Path, files)
			if err != nil {
				return nil, operationError(op.ID, "reading file: %v", err)
			}

			next, actual := insertBeforeLineBlock(current, op.FindLines, op.InsertLines)
			if actual != op.ExpectedOccurrences {
				return nil, operationError(op.ID, "expected %d occurrence(s), found %d", op.ExpectedOccurrences, actual)
			}

			state := files[op.Path]
			state.staged = next
			state.existsStaged = true
		case schema.ActionInsertAfterLines:
			current, err := stagedContent(root, op.Path, files)
			if err != nil {
				return nil, operationError(op.ID, "reading file: %v", err)
			}

			next, actual := insertAfterLineBlock(current, op.FindLines, op.InsertLines)
			if actual != op.ExpectedOccurrences {
				return nil, operationError(op.ID, "expected %d occurrence(s), found %d", op.ExpectedOccurrences, actual)
			}

			state := files[op.Path]
			state.staged = next
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

func replaceLineBlock(content string, findLines []string, replaceLines []string) (string, int) {
	return applyLineBlockOperation(content, findLines, replaceLines, lineBlockReplace)
}

func insertBeforeLineBlock(content string, findLines []string, insertLines []string) (string, int) {
	return applyLineBlockOperation(content, findLines, insertLines, lineBlockInsertBefore)
}

func insertAfterLineBlock(content string, findLines []string, insertLines []string) (string, int) {
	return applyLineBlockOperation(content, findLines, insertLines, lineBlockInsertAfter)
}

type lineBlockMode int

const (
	lineBlockReplace lineBlockMode = iota
	lineBlockInsertBefore
	lineBlockInsertAfter
)

func applyLineBlockOperation(content string, findLines []string, changeLines []string, mode lineBlockMode) (string, int) {
	lines, trailingNewline := splitTextLines(content)
	var next []string
	count := 0

	for i := 0; i < len(lines); {
		if hasLineBlockAt(lines, findLines, i) {
			switch mode {
			case lineBlockReplace:
				next = append(next, changeLines...)
			case lineBlockInsertBefore:
				next = append(next, changeLines...)
				next = append(next, findLines...)
			case lineBlockInsertAfter:
				next = append(next, findLines...)
				next = append(next, changeLines...)
			}
			i += len(findLines)
			count++
			continue
		}

		next = append(next, lines[i])
		i++
	}

	return joinTextLines(next, trailingNewline), count
}

func splitTextLines(content string) ([]string, bool) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	trailingNewline := strings.HasSuffix(normalized, "\n")
	if trailingNewline {
		normalized = strings.TrimSuffix(normalized, "\n")
	}
	if normalized == "" {
		if trailingNewline {
			return []string{""}, true
		}
		return nil, false
	}
	return strings.Split(normalized, "\n"), trailingNewline
}

func joinTextLines(lines []string, trailingNewline bool) string {
	content := strings.Join(lines, "\n")
	if trailingNewline {
		return content + "\n"
	}
	return content
}

func hasLineBlockAt(lines []string, findLines []string, start int) bool {
	if start+len(findLines) > len(lines) {
		return false
	}
	for i, findLine := range findLines {
		if lines[start+i] != findLine {
			return false
		}
	}
	return true
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

		builder.WriteString(diff.Unified(path, state.original, state.staged, state.existsOriginal, state.existsStaged))
	}
	return builder.String()
}
