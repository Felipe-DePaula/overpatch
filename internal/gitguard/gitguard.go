package gitguard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Sentinel errors for use with errors.Is.
var (
	ErrGitNotFound         = errors.New("git executable not found")
	ErrNotGitRepository    = errors.New("not inside a Git repository")
	ErrWorkingTreeNotClean = errors.New("Git working tree is not clean")
)

// Refusal is returned when apply is blocked by a Git precondition.
// It wraps a sentinel error (matchable with errors.Is) and carries a Hint for CLI output.
type Refusal struct {
	err  error
	Hint string
}

func (r *Refusal) Error() string { return r.err.Error() }
func (r *Refusal) Unwrap() error { return r.err }

func newRefusal(sentinel error, hint string) *Refusal {
	return &Refusal{err: sentinel, Hint: hint}
}

// CheckCleanWorkingTree verifies three preconditions before apply may write to disk:
//  1. git is available in PATH
//  2. root is inside a Git repository
//  3. the working tree is clean (no staged, unstaged, or untracked changes)
func CheckCleanWorkingTree(root string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return newRefusal(ErrGitNotFound,
			"install Git to use apply, or use validate/inspect/plan without writing changes")
	}

	out, err := gitCmd(root, "rev-parse", "--is-inside-work-tree").Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return newRefusal(ErrNotGitRepository,
			"run git init first, then commit or stash your current work before applying")
	}

	out, err = gitCmd(root, "status", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("checking git status: %w", err)
	}
	if strings.TrimSpace(string(out)) != "" {
		return newRefusal(ErrWorkingTreeNotClean,
			"commit or stash your changes before running apply")
	}

	return nil
}

// gitCmd builds a git command with Dir set to dir and git-specific environment
// variables stripped. Stripping prevents inherited GIT_DIR / GIT_WORK_TREE from
// the calling process (e.g. a test runner inside a repo) from overriding the
// working-directory detection that the check depends on.
func gitCmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	stripped := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		key := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			key = kv[:i]
		}
		switch key {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE",
			"GIT_OBJECT_DIRECTORY", "GIT_COMMON_DIR":
			// skip: these override repo detection and must not be inherited
		default:
			stripped = append(stripped, kv)
		}
	}
	cmd.Env = stripped
	return cmd
}
