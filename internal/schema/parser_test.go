package schema

import (
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		doc, err := ParseFile("../../examples/01-replace-text.json")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if doc == nil {
			t.Errorf("expected non-nil document")
		}
	})

	t.Run("BadSchemaVersion", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_bad_schema_version.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "schema_version") {
			t.Errorf("expected error to contain %q, got: %v", "schema_version", err)
		}
	})

	t.Run("SuccessNoOps", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_success_no_ops.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "status success") {
			t.Errorf("expected error to contain %q, got: %v", "status success", err)
		}
	})

	t.Run("DuplicateIDs", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_duplicate_ids.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("expected error to contain %q, got: %v", "duplicate", err)
		}
	})

	t.Run("UnknownAction", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_unknown_action.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "action invalid") {
			t.Errorf("expected error to contain %q, got: %v", "action invalid", err)
		}
	})
}
