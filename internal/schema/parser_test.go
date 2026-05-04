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

	t.Run("ReplaceTextMissingFind", func(t *testing.T) {
		_, err := ParseFile("testdata/replace_text_missing_find.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "find required for replace_text") {
			t.Errorf("expected error to contain %q, got: %v", "find required for replace_text", err)
		}
	})

	t.Run("ReplaceTextInvalidOccurrence", func(t *testing.T) {
		_, err := ParseFile("testdata/replace_text_invalid_occurrence.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "occurrence invalid for replace_text") {
			t.Errorf("expected error to contain %q, got: %v", "occurrence invalid for replace_text", err)
		}
	})

	t.Run("ReplaceTextMissingExpectedOccurrences", func(t *testing.T) {
		_, err := ParseFile("testdata/replace_text_missing_expected.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "expected_occurrences must be >= 1 for replace_text") {
			t.Errorf("expected error to contain %q, got: %v", "expected_occurrences must be >= 1 for replace_text", err)
		}
	})

	t.Run("ReplaceLinesMissingFindLines", func(t *testing.T) {
		_, err := ParseFile("testdata/replace_lines_missing_find.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "find_lines required for replace_lines") {
			t.Errorf("expected error to contain %q, got: %v", "find_lines required for replace_lines", err)
		}
	})

	t.Run("ReplaceLinesMissingReplaceLines", func(t *testing.T) {
		_, err := ParseFile("testdata/replace_lines_missing_replace.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "replace_lines required for replace_lines") {
			t.Errorf("expected error to contain %q, got: %v", "replace_lines required for replace_lines", err)
		}
	})

	t.Run("InsertAfterMissingInsertLines", func(t *testing.T) {
		_, err := ParseFile("testdata/insert_after_missing_insert.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "insert_lines required for insert_after_lines") {
			t.Errorf("expected error to contain %q, got: %v", "insert_lines required for insert_after_lines", err)
		}
	})

	t.Run("CreateMissingContent", func(t *testing.T) {
		_, err := ParseFile("testdata/create_missing_content.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "content required for create") {
			t.Errorf("expected error to contain %q, got: %v", "content required for create", err)
		}
	})

	t.Run("CreateEmptyContentValid", func(t *testing.T) {
		_, err := ParseFile("testdata/create_empty_content.json")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("DeleteValid", func(t *testing.T) {
		_, err := ParseFile("testdata/delete_valid.json")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("InvalidPathTraversal", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_path_traversal.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "path invalid") || !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("expected error to contain %q and %q, got: %v", "path invalid", "path traversal", err)
		}
	})

	t.Run("InvalidOperationIDNoPrefix", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_operation_id_no_prefix.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "id invalid") {
			t.Errorf("expected error to contain %q, got: %v", "id invalid", err)
		}
	})

	t.Run("InvalidOperationIDHyphen", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_operation_id_hyphen.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "id invalid") {
			t.Errorf("expected error to contain %q, got: %v", "id invalid", err)
		}
	})

	t.Run("InvalidOperationIDSpace", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_operation_id_space.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "id invalid") {
			t.Errorf("expected error to contain %q, got: %v", "id invalid", err)
		}
	})

	t.Run("InvalidOperationIDEmptySuffix", func(t *testing.T) {
		_, err := ParseFile("testdata/invalid_operation_id_empty_suffix.json")
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "id invalid") {
			t.Errorf("expected error to contain %q, got: %v", "id invalid", err)
		}
	})

	t.Run("ValidOperationIDWithLettersDigitsUnderscore", func(t *testing.T) {
		_, err := ParseFile("testdata/valid_operation_id_with_letters_digits_underscore.json")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}
