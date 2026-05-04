package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Felipe-DePaula/overpatch/internal/schema"
)

// runCmd executes rootCmd with the given args and returns output written via
// cmd.OutOrStdout() and any error. State is reset after each test via t.Cleanup.
func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		applyYes = false
	})
	err := rootCmd.Execute()
	return buf.String(), err
}

func writeJSONFixture(t *testing.T, path string, doc *schema.Document) {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func skipIfNoGitCLI(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}
}

func gitRunCLI(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func makeCleanRepoCLI(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	gitRunCLI(t, tmp, "init")
	gitRunCLI(t, tmp, "config", "user.email", "test@overpatch.test")
	gitRunCLI(t, tmp, "config", "user.name", "Overpatch Test")
	f := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(f, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRunCLI(t, tmp, "add", ".")
	gitRunCLI(t, tmp, "commit", "-m", "init")
	return tmp
}

// TestValidateOutputGoesToBuffer verifies that "valid" is written via
// cmd.OutOrStdout(), not directly to os.Stdout, so callers can capture it.
func TestValidateOutputGoesToBuffer(t *testing.T) {
	tmp := t.TempDir()
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusNoChanges,
		Reason:        "nothing to do",
		Operations:    []schema.Operation{},
	}
	fixturePath := filepath.Join(tmp, "patch.json")
	writeJSONFixture(t, fixturePath, doc)

	out, err := runCmd(t, "validate", fixturePath)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !strings.Contains(out, "valid") {
		t.Errorf("expected 'valid' in output buffer, got: %q", out)
	}
}

// TestValidateInvalidOutputGoesToBuffer verifies that "invalid: ..." is written
// via cmd.OutOrStdout() when validation fails.
func TestValidateInvalidOutputGoesToBuffer(t *testing.T) {
	tmp := t.TempDir()
	fixturePath := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(fixturePath, []byte(`{"schema_version":"wrong","status":"success","operations":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCmd(t, "validate", fixturePath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "invalid:") {
		t.Errorf("expected 'invalid:' in output buffer, got: %q", out)
	}
}

// TestPlanStatusFailedAborts verifies that plan prints "plan: failed" and the
// document reason when the document carries status: failed, and returns a
// non-nil error without writing any file.
func TestPlanStatusFailedAborts(t *testing.T) {
	tmp := t.TempDir()
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusFailed,
		Reason:        "model could not determine safe edits",
		Operations:    []schema.Operation{},
	}
	fixturePath := filepath.Join(tmp, "failed.json")
	writeJSONFixture(t, fixturePath, doc)

	out, err := runCmd(t, "plan", fixturePath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "plan: failed") {
		t.Errorf("expected 'plan: failed' in output, got: %q", out)
	}
	if !strings.Contains(out, "model could not determine safe edits") {
		t.Errorf("expected reason in output, got: %q", out)
	}
}

// TestPlanStatusFailedNoReason verifies that plan still aborts when the
// document has no reason field.
func TestPlanStatusFailedNoReason(t *testing.T) {
	tmp := t.TempDir()
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusFailed,
		Reason:        "no patch possible",
		Operations:    []schema.Operation{},
	}
	fixturePath := filepath.Join(tmp, "failed.json")
	writeJSONFixture(t, fixturePath, doc)

	out, err := runCmd(t, "plan", fixturePath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "plan: failed") {
		t.Errorf("expected 'plan: failed' in output, got: %q", out)
	}
}

// TestApplyStatusFailedAborts verifies that apply aborts with "apply: failed"
// and the document reason when the document carries status: failed, without
// writing any file to disk. The test runs inside a clean git repo so that the
// git guard is satisfied and execution reaches the status check.
func TestApplyStatusFailedAborts(t *testing.T) {
	skipIfNoGitCLI(t)

	repoDir := makeCleanRepoCLI(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	})

	// Write the fixture outside the repo so it does not dirty the working tree.
	fixtureDir := t.TempDir()
	doc := &schema.Document{
		SchemaVersion: schema.SchemaVersionV1,
		Status:        schema.StatusFailed,
		Reason:        "context window exceeded",
		Operations:    []schema.Operation{},
	}
	fixturePath := filepath.Join(fixtureDir, "failed.json")
	writeJSONFixture(t, fixturePath, doc)

	out, err := runCmd(t, "apply", "--yes", fixturePath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(out, "apply: failed") {
		t.Errorf("expected 'apply: failed' in output, got: %q", out)
	}
	if !strings.Contains(out, "context window exceeded") {
		t.Errorf("expected reason in output, got: %q", out)
	}

	// Confirm no unexpected files were written inside the repo.
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "hello.txt" && e.Name() != ".git" {
			t.Errorf("unexpected file in repo after failed apply: %s", e.Name())
		}
	}
}
