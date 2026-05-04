package gitguard_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Felipe-DePaula/overpatch/internal/gitguard"
)

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func makeCleanRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	gitRun(t, tmp, "init")
	gitRun(t, tmp, "config", "user.email", "test@overpatch.test")
	gitRun(t, tmp, "config", "user.name", "Overpatch Test")
	f := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(f, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, tmp, "add", ".")
	gitRun(t, tmp, "commit", "-m", "init")
	return tmp
}

func TestCheckCleanWorkingTreeOutsideRepo(t *testing.T) {
	skipIfNoGit(t)
	tmp := t.TempDir()
	err := gitguard.CheckCleanWorkingTree(tmp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, gitguard.ErrNotGitRepository) {
		t.Errorf("expected ErrNotGitRepository, got: %v", err)
	}
}

func TestCheckCleanWorkingTreeCleanRepo(t *testing.T) {
	skipIfNoGit(t)
	tmp := makeCleanRepo(t)
	if err := gitguard.CheckCleanWorkingTree(tmp); err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}
}

func TestCheckCleanWorkingTreeDirtyRepo(t *testing.T) {
	skipIfNoGit(t)
	tmp := makeCleanRepo(t)

	f := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(f, []byte("dirty\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := gitguard.CheckCleanWorkingTree(tmp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, gitguard.ErrWorkingTreeNotClean) {
		t.Errorf("expected ErrWorkingTreeNotClean, got: %v", err)
	}
}
