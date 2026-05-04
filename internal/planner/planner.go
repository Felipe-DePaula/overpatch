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

	originals := make(map[string]string)
	staged := make(map[string]string)

	for _, op := range doc.Operations {
		if op.Action != schema.ActionReplaceText {
			return nil, operationError(op.ID, "action not supported by plan yet: %s", op.Action)
		}

		current, err := stagedContent(root, op.Path, originals, staged)
		if err != nil {
			return nil, operationError(op.ID, "reading file: %v", err)
		}

		actual := strings.Count(current, op.Find)
		if actual != op.ExpectedOccurrences {
			return nil, operationError(op.ID, "expected %d occurrence(s), found %d", op.ExpectedOccurrences, actual)
		}

		next := replaceText(current, op.Find, op.Replace, op.Occurrence)
		staged[op.Path] = next
	}

	changedPaths := changedFilePaths(originals, staged)
	diff := renderDiff(originals, staged, changedPaths)

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

func stagedContent(root string, path string, originals map[string]string, staged map[string]string) (string, error) {
	if content, ok := staged[path]; ok {
		return content, nil
	}
	if content, ok := originals[path]; ok {
		return content, nil
	}

	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return "", err
	}

	content := string(data)
	originals[path] = content
	return content, nil
}

func replaceText(content string, find string, replace string, occurrence string) string {
	if occurrence == "first" {
		return strings.Replace(content, find, replace, 1)
	}
	return strings.ReplaceAll(content, find, replace)
}

func changedFilePaths(originals map[string]string, staged map[string]string) []string {
	paths := make([]string, 0, len(staged))
	for path, next := range staged {
		if originals[path] != next {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func renderDiff(originals map[string]string, staged map[string]string, paths []string) string {
	var builder strings.Builder
	for i, path := range paths {
		if i > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString("--- ")
		builder.WriteString(path)
		builder.WriteByte('\n')
		builder.WriteString("+++ ")
		builder.WriteString(path)
		builder.WriteByte('\n')
		builder.WriteString("@@\n")
		builder.WriteString(prefixLines("-", originals[path]))
		builder.WriteString(prefixLines("+", staged[path]))
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
