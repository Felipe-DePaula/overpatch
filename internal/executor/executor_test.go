package executor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Felipe-DePaula/overpatch/internal/planner"
)

func TestApplyModifiedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "file.txt",
			Kind:   planner.FileChangeModified,
			Staged: "new\n",
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	assertFileContent(t, path, "new\n")
}

func TestApplyModifiedFileDoesNotLeaveTempFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "file.txt",
			Kind:   planner.FileChangeModified,
			Staged: "new\n",
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	assertFileContent(t, path, "new\n")
	assertNoOverpatchTempFiles(t, root)
}

func TestApplyModifiedFilePreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not report POSIX file permissions reliably")
	}

	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("chmod fixture: %v", err)
	}

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "file.txt",
			Kind:   planner.FileChangeModified,
			Staged: "new\n",
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions = %v, want %v", got, os.FileMode(0o600))
	}
}

func TestApplyCreatedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "new.txt")

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "new.txt",
			Kind:   planner.FileChangeCreated,
			Staged: "created\n",
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	assertFileContent(t, path, "created\n")
}

func TestApplyCreatedFileDoesNotLeaveTempFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "new.txt")

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "new.txt",
			Kind:   planner.FileChangeCreated,
			Staged: "created\n",
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	assertFileContent(t, path, "created\n")
	assertNoOverpatchTempFiles(t, root)
}

func TestApplyCreatedFileUsesDefaultPermissions(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "new.txt")

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "new.txt",
			Kind:   planner.FileChangeCreated,
			Staged: "created\n",
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	assertFileContent(t, path, "created\n")
	assertNoOverpatchTempFiles(t, root)

	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Fatalf("permissions = %v, want %v", got, os.FileMode(0o644))
	}
}

func TestApplyCreatedFileWithMissingParentDirs(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "nested", "dir", "new.txt")

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "nested/dir/new.txt",
			Kind:   planner.FileChangeCreated,
			Staged: "created\n",
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	assertFileContent(t, path, "created\n")
}

func TestApplyDeletedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "old.txt")
	if err := os.WriteFile(path, []byte("remove\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path: "old.txt",
			Kind: planner.FileChangeDeleted,
		}},
	}

	if err := Apply(stage, root); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("deleted file still exists, stat error: %v", err)
	}
}

func TestApplyNilStageFails(t *testing.T) {
	err := Apply(nil, t.TempDir())
	if err == nil {
		t.Fatal("Apply returned nil error, want stage is nil")
	}
	if !strings.Contains(err.Error(), "stage is nil") {
		t.Fatalf("error = %q, want stage is nil", err)
	}
}

func TestApplyCreateExistingFileFails(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "existing.txt")
	if err := os.WriteFile(path, []byte("already here\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "existing.txt",
			Kind:   planner.FileChangeCreated,
			Staged: "new\n",
		}},
	}

	err := Apply(stage, root)
	if err == nil {
		t.Fatal("Apply returned nil error, want file already exists")
	}
	if !strings.Contains(err.Error(), "file already exists") {
		t.Fatalf("error = %q, want file already exists", err)
	}
}

func TestApplyDeleteDirectoryFails(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o700); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path: "docs",
			Kind: planner.FileChangeDeleted,
		}},
	}

	err := Apply(stage, root)
	if err == nil {
		t.Fatal("Apply returned nil error, want target is not a file")
	}
	if !strings.Contains(err.Error(), "target is not a file") {
		t.Fatalf("error = %q, want target is not a file", err)
	}
}

func TestApplyModifiedMissingFileFails(t *testing.T) {
	root := t.TempDir()

	stage := &planner.StageResult{
		Changes: []planner.FileChange{{
			Path:   "missing.txt",
			Kind:   planner.FileChangeModified,
			Staged: "new\n",
		}},
	}

	err := Apply(stage, root)
	if err == nil {
		t.Fatal("Apply returned nil error, want file does not exist")
	}
	if !strings.Contains(err.Error(), "file does not exist") {
		t.Fatalf("error = %q, want file does not exist", err)
	}
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != want {
		t.Fatalf("content = %q, want %q", string(data), want)
	}
}

func assertNoOverpatchTempFiles(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".overpatch-") {
			t.Fatalf("left temp file behind: %s", filepath.Join(dir, entry.Name()))
		}
	}
}
