// Package schema defines the Overpatch document model and validation logic.
package schema

// Document is the top-level structure of an Overpatch patch document.
type Document struct {
	SchemaVersion string      `json:"schema_version"`
	Status        string      `json:"status"`
	Reason        string      `json:"reason"`
	Summary       string      `json:"summary"`
	Operations    []Operation `json:"operations"`
}

// Operation describes a single deterministic patch action on a file.
type Operation struct {
	ID                  string   `json:"id"`
	Action              string   `json:"action"`
	Path                string   `json:"path"`
	Find                string   `json:"find,omitempty"`
	Replace             string   `json:"replace,omitempty"`
	Occurrence          string   `json:"occurrence,omitempty"`
	ExpectedOccurrences int      `json:"expected_occurrences,omitempty"`
	FindLines           []string `json:"find_lines,omitempty"`
	ReplaceLines        []string `json:"replace_lines,omitempty"`
	InsertLines         []string `json:"insert_lines,omitempty"`
	Content             *string  `json:"content,omitempty"`
}

// SchemaVersionV1 is the only supported schema version.
const SchemaVersionV1 = "overpatch/v1"

// Document status constants.
const (
	StatusSuccess   = "success"
	StatusNoChanges = "no_changes"
	StatusFailed    = "failed"
)

// Action type constants for Operation.Action.
const (
	ActionReplaceText       = "replace_text"
	ActionReplaceLines      = "replace_lines"
	ActionInsertBeforeLines = "insert_before_lines"
	ActionInsertAfterLines  = "insert_after_lines"
	ActionCreate            = "create"
	ActionDelete            = "delete"
)
