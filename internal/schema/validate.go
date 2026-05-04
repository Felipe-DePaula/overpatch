package schema

import (
	"fmt"
	"strings"

	"github.com/Felipe-DePaula/overpatch/internal/safety"
)

var allowedActions = map[string]bool{
	ActionReplaceText:       true,
	ActionReplaceLines:      true,
	ActionInsertBeforeLines: true,
	ActionInsertAfterLines:  true,
	ActionCreate:            true,
	ActionDelete:            true,
}

// ValidateDocument checks that doc conforms to the Overpatch v1 protocol.
// It performs structural validation, action-specific validation, and lexical
// path safety checks, but it does not inspect target file contents or resolve symlinks.
func ValidateDocument(doc *Document) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	if doc.SchemaVersion != SchemaVersionV1 {
		return fmt.Errorf("schema_version mismatch: got %q, want %q", doc.SchemaVersion, SchemaVersionV1)
	}

	switch doc.Status {
	case StatusSuccess, StatusNoChanges, StatusFailed:
	default:
		return fmt.Errorf("status invalid: %q", doc.Status)
	}

	if doc.Operations == nil {
		return fmt.Errorf("operations is nil")
	}

	switch doc.Status {
	case StatusSuccess:
		if len(doc.Operations) < 1 {
			return fmt.Errorf("status success requires at least one operation")
		}
	case StatusNoChanges, StatusFailed:
		if len(doc.Operations) != 0 {
			return fmt.Errorf("status %s requires empty operations", doc.Status)
		}
		if strings.TrimSpace(doc.Reason) == "" {
			return fmt.Errorf("status %s requires non-empty reason", doc.Status)
		}
	}

	seen := make(map[string]bool, len(doc.Operations))
	for i, op := range doc.Operations {
		if op.ID == "" {
			return fmt.Errorf("operation %d (id=%q): missing id", i, op.ID)
		}
		if !allowedActions[op.Action] {
			return fmt.Errorf("operation %s: action invalid: %q", op.ID, op.Action)
		}
		if op.Path == "" {
			return fmt.Errorf("operation %s: missing path", op.ID)
		}
		if err := safety.ValidateOperationPath(op.Path); err != nil {
			return fmt.Errorf("operation %s: path invalid: %v", op.ID, err)
		}
		if seen[op.ID] {
			return fmt.Errorf("duplicate operation id: %q", op.ID)
		}
		seen[op.ID] = true

		switch op.Action {
		case ActionReplaceText:
			if op.Find == "" {
				return fmt.Errorf("operation %s: find required for replace_text", op.ID)
			}
			if op.Occurrence != "all" && op.Occurrence != "first" {
				return fmt.Errorf("operation %s: occurrence invalid for replace_text: %q", op.ID, op.Occurrence)
			}
			if op.ExpectedOccurrences < 1 {
				return fmt.Errorf("operation %s: expected_occurrences must be >= 1 for replace_text", op.ID)
			}
		case ActionReplaceLines:
			if len(op.FindLines) == 0 {
				return fmt.Errorf("operation %s: find_lines required for replace_lines", op.ID)
			}
			if op.ReplaceLines == nil {
				return fmt.Errorf("operation %s: replace_lines required for replace_lines", op.ID)
			}
			if op.ExpectedOccurrences < 1 {
				return fmt.Errorf("operation %s: expected_occurrences must be >= 1 for replace_lines", op.ID)
			}
		case ActionInsertBeforeLines:
			if len(op.FindLines) == 0 {
				return fmt.Errorf("operation %s: find_lines required for insert_before_lines", op.ID)
			}
			if len(op.InsertLines) == 0 {
				return fmt.Errorf("operation %s: insert_lines required for insert_before_lines", op.ID)
			}
			if op.ExpectedOccurrences < 1 {
				return fmt.Errorf("operation %s: expected_occurrences must be >= 1 for insert_before_lines", op.ID)
			}
		case ActionInsertAfterLines:
			if len(op.FindLines) == 0 {
				return fmt.Errorf("operation %s: find_lines required for insert_after_lines", op.ID)
			}
			if len(op.InsertLines) == 0 {
				return fmt.Errorf("operation %s: insert_lines required for insert_after_lines", op.ID)
			}
			if op.ExpectedOccurrences < 1 {
				return fmt.Errorf("operation %s: expected_occurrences must be >= 1 for insert_after_lines", op.ID)
			}
		case ActionCreate:
			if op.Content == nil {
				return fmt.Errorf("operation %s: content required for create", op.ID)
			}
		case ActionDelete:
			// No extra fields required
		}
	}

	return nil
}
