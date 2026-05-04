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
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
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

func TestStageReplaceTextProducesModifiedChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	original := "Hello World\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stage, err := Stage(successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "README.md",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	}), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(stage.Changes))
	}

	change := stage.Changes[0]
	if change.Kind != FileChangeModified {
		t.Errorf("Kind = %q, want %q", change.Kind, FileChangeModified)
	}
	if !change.OriginalExists {
		t.Error("OriginalExists = false, want true")
	}
	if !change.StagedExists {
		t.Error("StagedExists = false, want true")
	}
	if !strings.Contains(change.Original, "Hello World") {
		t.Errorf("Original = %q, want old text", change.Original)
	}
	if !strings.Contains(change.Staged, "Ola Mundo") {
		t.Errorf("Staged = %q, want new text", change.Staged)
	}
	assertFileContent(t, path, original)
}

func TestStageCreateProducesCreatedChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "new.txt")
	content := "hello\n"

	stage, err := Stage(successDoc(schema.Operation{
		ID:      "op_001",
		Action:  schema.ActionCreate,
		Path:    "new.txt",
		Content: &content,
	}), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(stage.Changes))
	}

	change := stage.Changes[0]
	if change.Kind != FileChangeCreated {
		t.Errorf("Kind = %q, want %q", change.Kind, FileChangeCreated)
	}
	if change.OriginalExists {
		t.Error("OriginalExists = true, want false")
	}
	if !change.StagedExists {
		t.Error("StagedExists = false, want true")
	}
	if change.Staged != content {
		t.Errorf("Staged = %q, want %q", change.Staged, content)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("created file exists on disk after stage, stat error: %v", err)
	}
}

func TestStageDeleteProducesDeletedChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "old.txt")
	original := "remove me\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stage, err := Stage(successDoc(schema.Operation{
		ID:     "op_001",
		Action: schema.ActionDelete,
		Path:   "old.txt",
	}), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(stage.Changes))
	}

	change := stage.Changes[0]
	if change.Kind != FileChangeDeleted {
		t.Errorf("Kind = %q, want %q", change.Kind, FileChangeDeleted)
	}
	if !change.OriginalExists {
		t.Error("OriginalExists = false, want true")
	}
	if change.StagedExists {
		t.Error("StagedExists = true, want false")
	}
	if change.Original != original {
		t.Errorf("Original = %q, want %q", change.Original, original)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("deleted file should still exist on disk after stage: %v", err)
	}
}

func TestStageMultipleChangesSortedByPath(t *testing.T) {
	root := t.TempDir()
	fixtures := map[string]string{
		"z.txt": "z old\n",
		"a.txt": "a old\n",
	}
	for name, content := range fixtures {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}

	stage, err := Stage(successDoc(
		schema.Operation{
			ID:                  "op_001",
			Action:              schema.ActionReplaceText,
			Path:                "z.txt",
			Find:                "old",
			Replace:             "new",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:                  "op_002",
			Action:              schema.ActionReplaceText,
			Path:                "a.txt",
			Find:                "old",
			Replace:             "new",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
	), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 2 {
		t.Fatalf("len(Changes) = %d, want 2", len(stage.Changes))
	}
	if stage.Changes[0].Path != "a.txt" || stage.Changes[1].Path != "z.txt" {
		t.Fatalf("changes are not sorted by path: %#v", stage.Changes)
	}
}

func TestPlanStillRendersDiff(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("Hello World\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := Plan(successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "README.md",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	}), root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Diff == "" {
		t.Fatal("Diff is empty, want rendered diff")
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
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
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

func TestPlanInsertBeforeLinesSuccess(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	original := "before\nanchor 1\nanchor 2\nafter\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertBeforeLines,
		Path:                "README.md",
		FindLines:           []string{"anchor 1", "anchor 2"},
		InsertLines:         []string{"inserted before"},
		ExpectedOccurrences: 1,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
	}
	assertDiffOrder(t, result.Diff, "+inserted before\n", "+anchor 1\n")
	assertFileContent(t, path, original)
}

func TestPlanInsertAfterLinesSuccess(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	original := "before\nanchor 1\nanchor 2\nafter\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertAfterLines,
		Path:                "README.md",
		FindLines:           []string{"anchor 1", "anchor 2"},
		InsertLines:         []string{"inserted after"},
		ExpectedOccurrences: 1,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
	}
	assertDiffOrder(t, result.Diff, "+anchor 2\n", "+inserted after\n")
	assertFileContent(t, path, original)
}

func TestPlanInsertBeforeLinesMultipleOccurrences(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("anchor\nmiddle\nanchor\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertBeforeLines,
		Path:                "README.md",
		FindLines:           []string{"anchor"},
		InsertLines:         []string{"inserted before"},
		ExpectedOccurrences: 2,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if strings.Count(result.Diff, "+inserted before\n") != 2 {
		t.Errorf("Diff should contain both insertions:\n%s", result.Diff)
	}
}

func TestPlanInsertAfterLinesMultipleOccurrences(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("anchor\nmiddle\nanchor\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertAfterLines,
		Path:                "README.md",
		FindLines:           []string{"anchor"},
		InsertLines:         []string{"inserted after"},
		ExpectedOccurrences: 2,
	})

	result, err := Plan(doc, root)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if strings.Count(result.Diff, "+inserted after\n") != 2 {
		t.Errorf("Diff should contain both insertions:\n%s", result.Diff)
	}
}

func TestPlanInsertBeforeLinesOccurrenceMismatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("anchor\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertBeforeLines,
		Path:                "README.md",
		FindLines:           []string{"anchor"},
		InsertLines:         []string{"inserted before"},
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

func TestPlanInsertAfterLinesMissingFile(t *testing.T) {
	doc := successDoc(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertAfterLines,
		Path:                "README.md",
		FindLines:           []string{"anchor"},
		InsertLines:         []string{"inserted after"},
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
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
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
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
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
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
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

func TestPlanAllActionsSupported(t *testing.T) {
	root := t.TempDir()
	fixtures := map[string]string{
		"text.txt":   "hello\n",
		"lines.txt":  "old\nblock\n",
		"before.txt": "anchor\n",
		"after.txt":  "anchor\n",
		"delete.txt": "remove\n",
	}
	for name, content := range fixtures {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	content := "created\n"

	doc := successDoc(
		schema.Operation{
			ID:                  "op_001",
			Action:              schema.ActionReplaceText,
			Path:                "text.txt",
			Find:                "hello",
			Replace:             "hi",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:                  "op_002",
			Action:              schema.ActionReplaceLines,
			Path:                "lines.txt",
			FindLines:           []string{"old", "block"},
			ReplaceLines:        []string{"new", "block"},
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:                  "op_003",
			Action:              schema.ActionInsertBeforeLines,
			Path:                "before.txt",
			FindLines:           []string{"anchor"},
			InsertLines:         []string{"inserted before"},
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:                  "op_004",
			Action:              schema.ActionInsertAfterLines,
			Path:                "after.txt",
			FindLines:           []string{"anchor"},
			InsertLines:         []string{"inserted after"},
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:      "op_005",
			Action:  schema.ActionCreate,
			Path:    "create.txt",
			Content: &content,
		},
		schema.Operation{
			ID:     "op_006",
			Action: schema.ActionDelete,
			Path:   "delete.txt",
		},
	)

	result, err := Plan(doc, root)
	if err != nil {
		if strings.Contains(err.Error(), "action not supported by plan yet") {
			t.Fatalf("Plan returned unsupported action error: %v", err)
		}
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != schema.StatusSuccess {
		t.Fatalf("Status = %q, want %q", result.Status, schema.StatusSuccess)
	}
	if result.Operations != 6 {
		t.Errorf("Operations = %d, want 6", result.Operations)
	}
}

func TestStageCreateThenDeleteSamePathProducesNoChange(t *testing.T) {
	root := t.TempDir()
	content := "created\n"

	stage, err := Stage(successDoc(
		schema.Operation{
			ID:      "op_001",
			Action:  schema.ActionCreate,
			Path:    "x.txt",
			Content: &content,
		},
		schema.Operation{
			ID:     "op_002",
			Action: schema.ActionDelete,
			Path:   "x.txt",
		},
	), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 0 {
		t.Fatalf("len(Changes) = %d, want 0", len(stage.Changes))
	}
	if _, err := os.Stat(filepath.Join(root, "x.txt")); !os.IsNotExist(err) {
		t.Fatal("x.txt should not exist on disk")
	}
}

func TestStageDeleteThenCreateExistingFileProducesModifiedChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "x.txt")
	if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	newContent := "new\n"

	stage, err := Stage(successDoc(
		schema.Operation{
			ID:     "op_001",
			Action: schema.ActionDelete,
			Path:   "x.txt",
		},
		schema.Operation{
			ID:      "op_002",
			Action:  schema.ActionCreate,
			Path:    "x.txt",
			Content: &newContent,
		},
	), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(stage.Changes))
	}

	change := stage.Changes[0]
	if change.Kind != FileChangeModified {
		t.Errorf("Kind = %q, want %q", change.Kind, FileChangeModified)
	}
	if !change.OriginalExists {
		t.Error("OriginalExists = false, want true")
	}
	if !change.StagedExists {
		t.Error("StagedExists = false, want true")
	}
	if change.Original != "old\n" {
		t.Errorf("Original = %q, want %q", change.Original, "old\n")
	}
	if change.Staged != "new\n" {
		t.Errorf("Staged = %q, want %q", change.Staged, "new\n")
	}
	assertFileContent(t, path, "old\n")
}

func TestStageReplaceThenDeleteProducesDeletedChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "x.txt")
	if err := os.WriteFile(path, []byte("hello\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stage, err := Stage(successDoc(
		schema.Operation{
			ID:                  "op_001",
			Action:              schema.ActionReplaceText,
			Path:                "x.txt",
			Find:                "hello",
			Replace:             "bye",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:     "op_002",
			Action: schema.ActionDelete,
			Path:   "x.txt",
		},
	), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(stage.Changes))
	}

	change := stage.Changes[0]
	if change.Kind != FileChangeDeleted {
		t.Errorf("Kind = %q, want %q", change.Kind, FileChangeDeleted)
	}
	if !change.OriginalExists {
		t.Error("OriginalExists = false, want true")
	}
	if change.StagedExists {
		t.Error("StagedExists = true, want false")
	}
	assertFileContent(t, path, "hello\n")
}

func TestStageCreateThenReplaceTextProducesCreatedChange(t *testing.T) {
	root := t.TempDir()
	content := "hello\n"

	stage, err := Stage(successDoc(
		schema.Operation{
			ID:      "op_001",
			Action:  schema.ActionCreate,
			Path:    "x.txt",
			Content: &content,
		},
		schema.Operation{
			ID:                  "op_002",
			Action:              schema.ActionReplaceText,
			Path:                "x.txt",
			Find:                "hello",
			Replace:             "bye",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
	), root)
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if len(stage.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(stage.Changes))
	}

	change := stage.Changes[0]
	if change.Kind != FileChangeCreated {
		t.Errorf("Kind = %q, want %q", change.Kind, FileChangeCreated)
	}
	if change.Staged != "bye\n" {
		t.Errorf("Staged = %q, want %q", change.Staged, "bye\n")
	}
	if _, err := os.Stat(filepath.Join(root, "x.txt")); !os.IsNotExist(err) {
		t.Fatal("x.txt should not exist on disk")
	}
}

func TestStageDeleteDirectoryStillFails(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o700); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}

	_, err := Stage(successDoc(schema.Operation{
		ID:     "op_001",
		Action: schema.ActionDelete,
		Path:   "dir",
	}), root)
	if err == nil {
		t.Fatal("Stage returned nil error, want directory target error")
	}
	if !strings.Contains(err.Error(), "delete target is not a file") {
		t.Fatalf("error = %q, want delete target is not a file", err)
	}
}

func TestStageDeleteMissingStillFails(t *testing.T) {
	_, err := Stage(successDoc(schema.Operation{
		ID:     "op_001",
		Action: schema.ActionDelete,
		Path:   "missing.txt",
	}), t.TempDir())
	if err == nil {
		t.Fatal("Stage returned nil error, want missing file error")
	}
	if !strings.Contains(err.Error(), "file does not exist") {
		t.Fatalf("error = %q, want file does not exist", err)
	}
}

func TestStageStatusFailed(t *testing.T) {
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusFailed,
		Reason:        "model could not determine safe edits",
		Operations:    []schema.Operation{},
	}
	result, err := Stage(doc, t.TempDir())
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if result.Status != schema.StatusFailed {
		t.Errorf("Status = %q, want %q", result.Status, schema.StatusFailed)
	}
	if result.Reason != doc.Reason {
		t.Errorf("Reason = %q, want %q", result.Reason, doc.Reason)
	}
	if result.Operations != 0 {
		t.Errorf("Operations = %d, want 0", result.Operations)
	}
	if len(result.Changes) != 0 {
		t.Errorf("Changes = %d, want 0", len(result.Changes))
	}
}

func TestStageStatusFailedNoReason(t *testing.T) {
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusFailed,
		Reason:        "no patch possible",
		Operations:    []schema.Operation{},
	}
	result, err := Stage(doc, t.TempDir())
	if err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}
	if result.Status != schema.StatusFailed {
		t.Errorf("Status = %q, want %q", result.Status, schema.StatusFailed)
	}
}

func TestPlanStatusFailed(t *testing.T) {
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusFailed,
		Reason:        "context window exceeded",
		Operations:    []schema.Operation{},
	}
	result, err := Plan(doc, t.TempDir())
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if result.Status != schema.StatusFailed {
		t.Errorf("Status = %q, want %q", result.Status, schema.StatusFailed)
	}
	if result.Reason != doc.Reason {
		t.Errorf("Reason = %q, want %q", result.Reason, doc.Reason)
	}
	if result.Operations != 0 {
		t.Errorf("Operations = %d, want 0", result.Operations)
	}
}

func TestStageStatusFailedDoesNotTouchDisk(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "untouched.txt")
	original := "should not change\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusFailed,
		Reason:        "something went wrong",
		Operations:    []schema.Operation{},
	}
	if _, err := Stage(doc, root); err != nil {
		t.Fatalf("Stage returned error: %v", err)
	}

	assertFileContent(t, path, original)
}

func successDoc(ops ...schema.Operation) *schema.Document {
	return &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusSuccess,
		Reason:        "",
		Operations:    ops,
	}
}

func assertDiffOrder(t *testing.T, diff string, before string, after string) {
	t.Helper()

	beforeIndex := strings.Index(diff, before)
	if beforeIndex == -1 {
		t.Fatalf("Diff does not contain %q:\n%s", before, diff)
	}
	afterIndex := strings.Index(diff, after)
	if afterIndex == -1 {
		t.Fatalf("Diff does not contain %q:\n%s", after, diff)
	}
	if beforeIndex > afterIndex {
		t.Fatalf("Diff has %q after %q:\n%s", before, after, diff)
	}
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after plan: %v", err)
	}
	if string(data) != want {
		t.Fatalf("file was modified on disk: got %q, want %q", string(data), want)
	}
}
