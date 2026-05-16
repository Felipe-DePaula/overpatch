package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Felipe-DePaula/overpatch/internal/schema"
)

// ---------------------------------------------------------------------------
// Helpers shared across integration tests
// ---------------------------------------------------------------------------

func ptr(s string) *string { return &s }

func makeDoc(status string, reason string, ops ...schema.Operation) *schema.Document {
	return &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        status,
		Reason:        reason,
		Operations:    ops,
	}
}

func successOps(ops ...schema.Operation) *schema.Document {
	return makeDoc(schema.StatusSuccess, "", ops...)
}

// writeFixtureFile writes content to name inside dir, fails the test on error.
func writeFixtureFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
	return p
}

// readFixtureFile reads the content of name inside dir.
func readFixtureFile(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

// cdAndRestore changes the working directory to dir and restores it on cleanup.
func cdAndRestore(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// inspect
// ---------------------------------------------------------------------------

func TestIntegrationInspectValidDoc(t *testing.T) {
	tmp := t.TempDir()
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusNoChanges,
		Reason:        "nothing to do",
		Summary:       "test summary",
		Operations:    []schema.Operation{},
	}
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "inspect", fix)
	if err != nil {
		t.Fatalf("inspect returned error: %v", err)
	}
	if !strings.Contains(out, "schema: overpatch/v1") {
		t.Errorf("output missing schema line: %q", out)
	}
	if !strings.Contains(out, "status: no_changes") {
		t.Errorf("output missing status line: %q", out)
	}
	if !strings.Contains(out, "summary: test summary") {
		t.Errorf("output missing summary line: %q", out)
	}
	if !strings.Contains(out, "reason: nothing to do") {
		t.Errorf("output missing reason line: %q", out)
	}
	if !strings.Contains(out, "operations: 0") {
		t.Errorf("output missing operations line: %q", out)
	}
}

func TestIntegrationInspectSuccessDocListsOps(t *testing.T) {
	tmp := t.TempDir()
	content := "hello\n"
	doc := successOps(
		schema.Operation{
			ID:                  "op_001",
			Action:              schema.ActionReplaceText,
			Path:                "file.txt",
			Find:                "x",
			Replace:             "y",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:      "op_002",
			Action:  schema.ActionCreate,
			Path:    "new.txt",
			Content: &content,
		},
	)
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "inspect", fix)
	if err != nil {
		t.Fatalf("inspect returned error: %v", err)
	}
	if !strings.Contains(out, "operations: 2") {
		t.Errorf("output missing operations count: %q", out)
	}
	if !strings.Contains(out, "op_001 replace_text file.txt") {
		t.Errorf("output missing op_001 line: %q", out)
	}
	if !strings.Contains(out, "op_002 create new.txt") {
		t.Errorf("output missing op_002 line: %q", out)
	}
}

func TestIntegrationInspectInvalidDoc(t *testing.T) {
	tmp := t.TempDir()
	fix := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(fix, []byte(`{"schema_version":"bad","status":"success","operations":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, "inspect", fix)
	if err == nil {
		t.Fatal("inspect returned nil error, want error")
	}
	if !strings.Contains(out, "invalid:") {
		t.Errorf("output missing invalid prefix: %q", out)
	}
}

// ---------------------------------------------------------------------------
// validate — security cases
// ---------------------------------------------------------------------------

func TestIntegrationValidateRejectsPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "../outside.txt",
		Find:                "x",
		Replace:             "y",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "validate", fix)
	if err == nil {
		t.Fatal("validate returned nil error, want path traversal error")
	}
	if !strings.Contains(out, "invalid:") {
		t.Errorf("output missing invalid prefix: %q", out)
	}
}

func TestIntegrationValidateRejectsAbsolutePath(t *testing.T) {
	tmp := t.TempDir()
	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "/etc/passwd",
		Find:                "root",
		Replace:             "owned",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "validate", fix)
	if err == nil {
		t.Fatal("validate returned nil error, want absolute path error")
	}
	if !strings.Contains(out, "invalid:") {
		t.Errorf("output missing invalid prefix: %q", out)
	}
}

func TestIntegrationValidateRejectsDotGitPath(t *testing.T) {
	tmp := t.TempDir()
	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                ".git/config",
		Find:                "x",
		Replace:             "y",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "validate", fix)
	if err == nil {
		t.Fatal("validate returned nil error, want blocked path error")
	}
	if !strings.Contains(out, "invalid:") {
		t.Errorf("output missing invalid prefix: %q", out)
	}
}

func TestIntegrationValidateRejectsDotEnvPath(t *testing.T) {
	tmp := t.TempDir()
	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                ".env",
		Find:                "x",
		Replace:             "y",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "validate", fix)
	if err == nil {
		t.Fatal("validate returned nil error, want blocked path error")
	}
	if !strings.Contains(out, "invalid:") {
		t.Errorf("output missing invalid prefix: %q", out)
	}
}

// ---------------------------------------------------------------------------
// plan — does not write to disk
// ---------------------------------------------------------------------------

func TestIntegrationPlanReplaceTextNoWrite(t *testing.T) {
	tmp := t.TempDir()
	writeFixtureFile(t, tmp, "hello.txt", "Hello World\n")

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "hello.txt",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	cdAndRestore(t, tmp)

	out, err := runCmd(t, "plan", fix)
	if err != nil {
		t.Fatalf("plan returned error: %v", err)
	}
	if !strings.Contains(out, "plan: success") {
		t.Errorf("output missing plan: success: %q", out)
	}
	if !strings.Contains(out, "Hello World") || !strings.Contains(out, "Ola Mundo") {
		t.Errorf("output missing diff content: %q", out)
	}

	if got := readFixtureFile(t, tmp, "hello.txt"); got != "Hello World\n" {
		t.Errorf("plan wrote to disk: got %q", got)
	}
}

func TestIntegrationPlanAnchorNotFound(t *testing.T) {
	tmp := t.TempDir()
	writeFixtureFile(t, tmp, "hello.txt", "Hello World\n")

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "hello.txt",
		Find:                "DOES NOT EXIST",
		Replace:             "nope",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	cdAndRestore(t, tmp)

	out, err := runCmd(t, "plan", fix)
	if err == nil {
		t.Fatal("plan returned nil error, want anchor/occurrence error")
	}
	if !strings.Contains(out, "plan: failed") {
		t.Errorf("output missing plan: failed: %q", out)
	}
}

func TestIntegrationPlanOccurrenceMismatch(t *testing.T) {
	tmp := t.TempDir()
	writeFixtureFile(t, tmp, "hello.txt", "Hello World\n")

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "hello.txt",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 5,
	})
	fix := filepath.Join(tmp, "doc.json")
	writeJSONFixture(t, fix, doc)

	cdAndRestore(t, tmp)

	out, err := runCmd(t, "plan", fix)
	if err == nil {
		t.Fatal("plan returned nil error, want occurrence mismatch")
	}
	if !strings.Contains(out, "plan: failed") {
		t.Errorf("output missing plan: failed: %q", out)
	}
}

// ---------------------------------------------------------------------------
// apply — requires --yes
// ---------------------------------------------------------------------------

func TestIntegrationApplyRequiresYes(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	cdAndRestore(t, repoDir)

	doc := makeDoc(schema.StatusNoChanges, "nothing to do")
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", fix)
	if err == nil {
		t.Fatal("apply returned nil error without --yes")
	}
	if !strings.Contains(out, "apply: refused") {
		t.Errorf("output missing apply: refused: %q", out)
	}
}

// ---------------------------------------------------------------------------
// apply — dirty working tree is refused
// ---------------------------------------------------------------------------

func TestIntegrationApplyRefusesDirtyTree(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)

	// Dirty the tree with an untracked file.
	if err := os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cdAndRestore(t, repoDir)

	fixtureDir := t.TempDir()
	doc := makeDoc(schema.StatusNoChanges, "nothing to do")
	fix := filepath.Join(fixtureDir, "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err == nil {
		t.Fatal("apply returned nil error on dirty tree")
	}
	if !strings.Contains(out, "apply: refused") {
		t.Errorf("output missing apply: refused: %q", out)
	}
}

// ---------------------------------------------------------------------------
// apply — happy path for each operation type
// ---------------------------------------------------------------------------

func TestIntegrationApplyReplaceText(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "greet.txt", "Hello World\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add greet")

	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "greet.txt",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err != nil {
		t.Fatalf("apply returned error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "apply: success") {
		t.Errorf("output missing apply: success: %q", out)
	}
	if got := readFixtureFile(t, repoDir, "greet.txt"); got != "Ola Mundo\n" {
		t.Errorf("file content = %q, want %q", got, "Ola Mundo\n")
	}
}

func TestIntegrationApplyReplaceLines(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "code.txt", "before\nold line 1\nold line 2\nafter\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add code")

	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceLines,
		Path:                "code.txt",
		FindLines:           []string{"old line 1", "old line 2"},
		ReplaceLines:        []string{"new line 1", "new line 2"},
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err != nil {
		t.Fatalf("apply returned error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "apply: success") {
		t.Errorf("output missing apply: success: %q", out)
	}
	got := readFixtureFile(t, repoDir, "code.txt")
	if !strings.Contains(got, "new line 1") {
		t.Errorf("file content = %q, expected new lines", got)
	}
	if strings.Contains(got, "old line 1") {
		t.Errorf("file content = %q, old lines still present", got)
	}
}

func TestIntegrationApplyInsertBeforeLines(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "src.txt", "anchor line\nend\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add src")

	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertBeforeLines,
		Path:                "src.txt",
		FindLines:           []string{"anchor line"},
		InsertLines:         []string{"inserted before"},
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err != nil {
		t.Fatalf("apply returned error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "apply: success") {
		t.Errorf("output missing apply: success: %q", out)
	}
	got := readFixtureFile(t, repoDir, "src.txt")
	idx1 := strings.Index(got, "inserted before")
	idx2 := strings.Index(got, "anchor line")
	if idx1 == -1 || idx2 == -1 || idx1 > idx2 {
		t.Errorf("inserted line not before anchor in: %q", got)
	}
}

func TestIntegrationApplyInsertAfterLines(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "src.txt", "anchor line\nend\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add src")

	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionInsertAfterLines,
		Path:                "src.txt",
		FindLines:           []string{"anchor line"},
		InsertLines:         []string{"inserted after"},
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err != nil {
		t.Fatalf("apply returned error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "apply: success") {
		t.Errorf("output missing apply: success: %q", out)
	}
	got := readFixtureFile(t, repoDir, "src.txt")
	idx1 := strings.Index(got, "anchor line")
	idx2 := strings.Index(got, "inserted after")
	if idx1 == -1 || idx2 == -1 || idx1 > idx2 {
		t.Errorf("inserted line not after anchor in: %q", got)
	}
}

func TestIntegrationApplyCreateFile(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:      "op_001",
		Action:  schema.ActionCreate,
		Path:    "brand-new.txt",
		Content: ptr("created content\n"),
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err != nil {
		t.Fatalf("apply returned error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "apply: success") {
		t.Errorf("output missing apply: success: %q", out)
	}
	got := readFixtureFile(t, repoDir, "brand-new.txt")
	if got != "created content\n" {
		t.Errorf("created file content = %q, want %q", got, "created content\n")
	}
}

func TestIntegrationApplyDeleteFile(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "remove-me.txt", "goodbye\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add file to delete")

	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:     "op_001",
		Action: schema.ActionDelete,
		Path:   "remove-me.txt",
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err != nil {
		t.Fatalf("apply returned error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "apply: success") {
		t.Errorf("output missing apply: success: %q", out)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "remove-me.txt")); !os.IsNotExist(err) {
		t.Error("deleted file still exists on disk after apply")
	}
}

// ---------------------------------------------------------------------------
// apply — failure cases at staging
// ---------------------------------------------------------------------------

func TestIntegrationApplyAnchorNotFound(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "greet.txt", "Hello World\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add greet")

	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "greet.txt",
		Find:                "MISSING ANCHOR",
		Replace:             "irrelevant",
		Occurrence:          "all",
		ExpectedOccurrences: 1,
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err == nil {
		t.Fatal("apply returned nil error, want anchor not found error")
	}
	if !strings.Contains(out, "apply: failed") {
		t.Errorf("output missing apply: failed: %q", out)
	}
	if got := readFixtureFile(t, repoDir, "greet.txt"); got != "Hello World\n" {
		t.Errorf("file was modified after failed apply: %q", got)
	}
}

func TestIntegrationApplyOccurrenceMismatch(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "greet.txt", "Hello World\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add greet")

	cdAndRestore(t, repoDir)

	doc := successOps(schema.Operation{
		ID:                  "op_001",
		Action:              schema.ActionReplaceText,
		Path:                "greet.txt",
		Find:                "Hello World",
		Replace:             "Ola Mundo",
		Occurrence:          "all",
		ExpectedOccurrences: 99,
	})
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err == nil {
		t.Fatal("apply returned nil error, want occurrence mismatch error")
	}
	if !strings.Contains(out, "apply: failed") {
		t.Errorf("output missing apply: failed: %q", out)
	}
	if got := readFixtureFile(t, repoDir, "greet.txt"); got != "Hello World\n" {
		t.Errorf("file was modified after failed apply: %q", got)
	}
}

// TestIntegrationApplyAtomicAbortOnFailure verifies that when a batch contains
// a valid first operation and an invalid second operation, the valid operation
// is NOT written because staging fails before any disk write.
func TestIntegrationApplyAtomicAbortOnFailure(t *testing.T) {
	skipIfNoGitCLI(t)
	repoDir := makeCleanRepoCLI(t)
	writeFixtureFile(t, repoDir, "a.txt", "original a\n")
	writeFixtureFile(t, repoDir, "b.txt", "original b\n")
	gitRunCLI(t, repoDir, "add", ".")
	gitRunCLI(t, repoDir, "commit", "-m", "add files")

	cdAndRestore(t, repoDir)

	doc := successOps(
		schema.Operation{
			ID:                  "op_001",
			Action:              schema.ActionReplaceText,
			Path:                "a.txt",
			Find:                "original a",
			Replace:             "modified a",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
		schema.Operation{
			ID:                  "op_002",
			Action:              schema.ActionReplaceText,
			Path:                "b.txt",
			Find:                "ANCHOR NOT PRESENT",
			Replace:             "irrelevant",
			Occurrence:          "all",
			ExpectedOccurrences: 1,
		},
	)
	fix := filepath.Join(t.TempDir(), "doc.json")
	writeJSONFixture(t, fix, doc)

	out, err := runCmd(t, "apply", "--yes", fix)
	if err == nil {
		t.Fatal("apply returned nil error, want staging failure")
	}
	if !strings.Contains(out, "apply: failed") {
		t.Errorf("output missing apply: failed: %q", out)
	}
	if got := readFixtureFile(t, repoDir, "a.txt"); got != "original a\n" {
		t.Errorf("a.txt was modified despite staging failure: %q", got)
	}
}
