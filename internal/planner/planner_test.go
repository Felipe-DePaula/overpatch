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
		Action:              schema.ActionInsertAfterLines,
		Path:                "README.md",
		FindLines:           []string{"old"},
		InsertLines:         []string{"new"},
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

func TestPlanReplaceLinesSuccess(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	original := "before\nold line 1\nold line 2\nafter\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceLines,
		Path:                "README.md",
		FindLines:           []string{"old line 1", "old line 2"},
		ReplaceLines:        []string{"new line 1", "new line 2"},
		ExpectedOccurrences: 1,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, StatusSuccess)
	}
	if !strings.Contains(result.Diff, "old line 1") {
		t.Errorf("Diff should contain old line:\n%s", result.Diff)
	}
	if !strings.Contains(result.Diff, "new line 1") {
		t.Errorf("Diff should contain new line:\n%s", result.Diff)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after plan: %v", err)
	}
	if string(data) != original {
		t.Errorf("file was modified on disk: got %q, want %q", string(data), original)
	}
}

func TestPlanReplaceLinesRemoveBlock(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	original := "before\nremove line 1\nremove line 2\nafter\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceLines,
		Path:                "README.md",
		FindLines:           []string{"remove line 1", "remove line 2"},
		ReplaceLines:        []string{},
		ExpectedOccurrences: 1,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if !strings.Contains(result.Diff, "-remove line 1") {
		t.Errorf("Diff should show removed block:\n%s", result.Diff)
	}
	if strings.Contains(result.Diff, "+remove line 1") {
		t.Errorf("Diff should not keep removed block in staged content:\n%s", result.Diff)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after plan: %v", err)
	}
	if string(data) != original {
		t.Errorf("file was modified on disk: got %q, want %q", string(data), original)
	}
}

func TestPlanReplaceLinesMultipleOccurrences(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("old\nblock\nmiddle\nold\nblock\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceLines,
		Path:                "README.md",
		FindLines:           []string{"old", "block"},
		ReplaceLines:        []string{"new", "block"},
		ExpectedOccurrences: 2,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if strings.Count(result.Diff, "+new") != 2 {
		t.Errorf("Diff should contain both replacements:\n%s", result.Diff)
	}
}

func TestPlanReplaceLinesOccurrenceMismatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("old line 1\nold line 2\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceLines,
		Path:                "README.md",
		FindLines:           []string{"old line 1", "old line 2"},
		ReplaceLines:        []string{"new line 1"},
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

func TestPlanReplaceLinesMissingFile(t *testing.T) {
	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceLines,
		Path:                "README.md",
		FindLines:           []string{"old line 1", "old line 2"},
		ReplaceLines:        []string{"new line 1"},
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

func TestPlanCreateFileSuccess(t *testing.T) {
	root := t.TempDir()
	content := "hello\n"
	path := filepath.Join(root, "new.txt")

	doc := successDoc(schema.Operation{
		ID:      "op_001",
		Action:  schema.ActionCreate,
		Path:    "new.txt",
		Content: &content,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, StatusSuccess)
	}
	if result.FilesChanged != 1 {
		t.Errorf("FilesChanged = %d, want 1", result.FilesChanged)
	}
	if !strings.Contains(result.Diff, "/dev/null") {
		t.Errorf("Diff should indicate a new file:\n%s", result.Diff)
	}
	if !strings.Contains(result.Diff, "+hello") {
		t.Errorf("Diff should contain created content:\n%s", result.Diff)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("created file exists on disk after plan, stat error: %v", err)
	}
}

func TestPlanCreateEmptyFileSuccess(t *testing.T) {
	root := t.TempDir()
	content := ""
	path := filepath.Join(root, "empty.txt")

	doc := successDoc(schema.Operation{
		ID:      "op_001",
		Action:  schema.ActionCreate,
		Path:    "empty.txt",
		Content: &content,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, StatusSuccess)
	}
	if result.FilesChanged != 1 {
		t.Errorf("FilesChanged = %d, want 1", result.FilesChanged)
	}
	if !strings.Contains(result.Diff, "--- /dev/null") || !strings.Contains(result.Diff, "+++ empty.txt") {
		t.Errorf("Diff should indicate an empty new file:\n%s", result.Diff)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("created file exists on disk after plan, stat error: %v", err)
	}
}

func TestPlanCreateFileAlreadyExists(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "existing.txt"), []byte("already here\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	content := "new content\n"

	doc := successDoc(schema.Operation{
		ID:      "op_001",
		Action:  schema.ActionCreate,
		Path:    "existing.txt",
		Content: &content,
	})

	_, err := Plan(doc, root)
	if err == nil {
		t.Fatal("Plan returned nil error, want file already exists error")
	}
	if !strings.Contains(err.Error(), "file already exists") {
		t.Fatalf("error = %q, want file already exists", err)
	}
}

func TestPlanDeleteFileSuccess(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "old.txt")
	original := "remove me\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:     "op_001",
		Action: schema.ActionDelete,
		Path:   "old.txt",
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, StatusSuccess)
	}
	if result.FilesChanged != 1 {
		t.Errorf("FilesChanged = %d, want 1", result.FilesChanged)
	}
	if !strings.Contains(result.Diff, "+++ /dev/null") {
		t.Errorf("Diff should indicate deletion:\n%s", result.Diff)
	}
	if !strings.Contains(result.Diff, "-remove me") {
		t.Errorf("Diff should contain removed content:\n%s", result.Diff)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("deleted file should still exist on disk after plan: %v", err)
	}
}

func TestPlanDeleteMissingFile(t *testing.T) {
	doc := successDoc(schema.Operation{
		ID:     "op_001",
		Action: schema.ActionDelete,
		Path:   "missing.txt",
	})

	_, err := Plan(doc, t.TempDir())
	if err == nil {
		t.Fatal("Plan returned nil error, want missing file error")
	}
	if !strings.Contains(err.Error(), "file does not exist") {
		t.Fatalf("error = %q, want file does not exist", err)
	}
}

func TestPlanDeleteDirectoryFails(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o700); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:     "op_001",
		Action: schema.ActionDelete,
		Path:   "docs",
	})

	_, err := Plan(doc, root)
	if err == nil {
		t.Fatal("Plan returned nil error, want directory target error")
	}
	if !strings.Contains(err.Error(), "delete target is not a file") {
		t.Fatalf("error = %q, want delete target is not a file", err)
	}
}

func TestPlanUnsupportedInsertBeforeStillFails(t *testing.T) {
	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertBeforeLines,
		Path:                "README.md",
		FindLines:           []string{"old"},
		InsertLines:         []string{"new"},
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

func successDoc(ops ...schema.Operation) *schema.Document {
	return &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusSuccess,
		Reason:        "",
		Operations:    ops,
	}
}
