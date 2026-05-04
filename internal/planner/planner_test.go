package planner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Felipe-DePaula/overpatch/internal/schema"
)

func TestPlanReplaceTextAllSuccess(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	original := "Hello World\nAgain: Hello World\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "README.md",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 2,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, StatusSuccess)
	}
	if result.Operations != 1 {
		t.Errorf("Operations = %d, want 1", result.Operations)
	}
	if result.FilesChanged != 1 {
		t.Errorf("FilesChanged = %d, want 1", result.FilesChanged)
	}
	if !strings.Contains(result.Diff, "Hello World") || !strings.Contains(result.Diff, "Ola Mundo") {
		t.Errorf("Diff does not contain old and new text:\n%s", result.Diff)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after plan: %v", err)
	}
	if string(data) != original {
		t.Errorf("file was modified on disk: got %q, want %q", string(data), original)
	}
}

func TestPlanReplaceTextFirstSuccess(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("one one\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "README.md",
		Find:                "one",
		Replace:             "two",
		Occurrence:          "first",
		ExpectedOccurrences: 2,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if !strings.Contains(result.Diff, "+two one") {
		t.Errorf("Diff should show only the first occurrence replaced:\n%s", result.Diff)
	}

	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read fixture after plan: %v", err)
	}
	if string(data) != "one one\n" {
		t.Errorf("file was modified on disk: got %q", string(data))
	}
}

func TestPlanReplaceTextOccurrenceMismatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("Hello World\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "README.md",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 2,
	})

	_, err := Plan(doc, root)
	if err == nil {
		t.Fatal("Plan returned nil error, want mismatch error")
	}
	if !strings.Contains(err.Error(), "expected 2 occurrence(s), found 1") {
		t.Fatalf("error = %q, want occurrence mismatch", err)
	}
}

func TestPlanUnsupportedAction(t *testing.T) {
	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceLines,
		Path:                "README.md",
		FindLines:           []string{"old"},
		ReplaceLines:        []string{"new"},
		ExpectedOccurrences: 1,
	})

	_, err := Plan(doc, t.TempDir())
	if err == nil {
		t.Fatal("Plan returned nil error, want unsupported action error")
	}
	if !strings.Contains(err.Error(), "action not supported by plan yet") {
		t.Fatalf("error = %q, want unsupported action", err)
	}
}

func TestPlanMissingFile(t *testing.T) {
	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "README.md",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})

	_, err := Plan(doc, t.TempDir())
	if err == nil {
		t.Fatal("Plan returned nil error, want missing file error")
	}
	if !strings.Contains(err.Error(), "reading file") {
		t.Fatalf("error = %q, want reading file", err)
	}
}

func successDoc(ops ...schema.Operation) *schema.Document {
	return &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusSuccess,
		Reason:        "",
		Operations:    ops,
	}
}
