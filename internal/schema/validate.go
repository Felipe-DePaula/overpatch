package schema

import (
	"fmt"
	"strings"
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
// It performs structural validation only; it does not inspect file paths or
// action-specific field requirements.
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
		if seen[op.ID] {
			return fmt.Errorf("duplicate operation id: %q", op.ID)
		}
		seen[op.ID] = true
	}

	return nil
}
