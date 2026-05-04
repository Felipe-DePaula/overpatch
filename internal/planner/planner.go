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

// FileChangeKind classifies how a staged file differs from disk.
type FileChangeKind string

const (
	FileChangeModified FileChangeKind = "modified"
	FileChangeCreated  FileChangeKind = "created"
	FileChangeDeleted  FileChangeKind = "deleted"
)

// FileChange describes the original and staged content for a single path.
type FileChange struct {
	Path           string
	Kind           FileChangeKind
	Original       string
	Staged         string
	OriginalExists bool
	StagedExists   bool
}

// StageResult describes the outcome of applying operations in memory.
type StageResult struct {
	Status     string
	Reason     string
	Operations int
	Changes    []FileChange
}

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
	isDir          bool
	loaded         bool
}

// Stage validates staging requirements and applies supported operations to
// in-memory file contents without writing to disk.
func Stage(doc *schema.Document, root string) (*StageResult, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	if doc.Status == schema.StatusNoChanges {
		return &StageResult{
			Status:     schema.StatusNoChanges,
			Reason:     doc.Reason,
			Operations: 0,
		}, nil
	}

	if doc.Status == schema.StatusFailed {
		return &StageResult{
			Status:     schema.StatusFailed,
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

	changes := collectFileChanges(files)

	return &StageResult{
		Status:     schema.StatusSuccess,
		Operations: len(doc.Operations),
		Changes:    changes,
	}, nil
}

// Plan stages supported operations in memory and returns a textual preview.
func Plan(doc *schema.Document, root string) (*Result, error) {
	stage, err := Stage(doc, root)
	if err != nil {
		return nil, err
	}

	if stage.Status == schema.StatusNoChanges {
		return &Result{
			Status:     schema.StatusNoChanges,
			Reason:     stage.Reason,
			Operations: 0,
		}, nil
	}

	if stage.Status == schema.StatusFailed {
		return &Result{
			Status:     schema.StatusFailed,
			Reason:     stage.Reason,
			Operations: 0,
		}, nil
	}

	return &Result{
		Status:       schema.StatusSuccess,
		Operations:   stage.Operations,
		FilesChanged: len(stage.Changes),
		Diff:         renderDiff(stage.Changes),
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
	if state.isDir {
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
		state.isDir = true
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

func collectFileChanges(files map[string]*fileState) []FileChange {
	changes := make([]FileChange, 0, len(files))
	for path, state := range files {
		if state.existsOriginal != state.existsStaged || state.original != state.staged {
			changes = append(changes, FileChange{
				Path:           path,
				Kind:           fileChangeKind(state),
				Original:       state.original,
				Staged:         state.staged,
				OriginalExists: state.existsOriginal,
				StagedExists:   state.existsStaged,
			})
		}
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})
	return changes
}

func fileChangeKind(state *fileState) FileChangeKind {
	if !state.existsOriginal && state.existsStaged {
		return FileChangeCreated
	}
	if state.existsOriginal && !state.existsStaged {
		return FileChangeDeleted
	}
	return FileChangeModified
}

func renderDiff(changes []FileChange) string {
	var builder strings.Builder
	for i, change := range changes {
		if i > 0 {
			builder.WriteByte('\n')
		}

		builder.WriteString(diff.Unified(
			change.Path,
			change.Original,
			change.Staged,
			change.OriginalExists,
			change.StagedExists,
		))
	}
	return builder.String()
}
