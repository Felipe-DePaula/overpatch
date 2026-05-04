package diff

import (
	"strings"
	"testing"
)

func TestUnifiedModifiedFileDiff(t *testing.T) {
	got := Unified("file.txt", "old\n", "new\n", true, true)

	assertContains(t, got, "--- file.txt")
	assertContains(t, got, "+++ file.txt")
	assertContains(t, got, "-old")
	assertContains(t, got, "+new")
}

func TestUnifiedCreatedFileDiff(t *testing.T) {
	got := Unified("file.txt", "", "hello\n", false, true)

	assertContains(t, got, "--- /dev/null")
	assertContains(t, got, "+++ file.txt")
	assertContains(t, got, "+hello")
}

func TestUnifiedDeletedFileDiff(t *testing.T) {
	got := Unified("file.txt", "bye\n", "", true, false)

	assertContains(t, got, "--- file.txt")
	assertContains(t, got, "+++ /dev/null")
	assertContains(t, got, "-bye")
}

func TestUnifiedCreatedEmptyFileDiff(t *testing.T) {
	got := Unified("file.txt", "", "", false, true)

	assertContains(t, got, "--- /dev/null")
	assertContains(t, got, "+++ file.txt")
	assertContains(t, got, "@@")
}

func TestUnifiedDeletedEmptyFileDiff(t *testing.T) {
	got := Unified("file.txt", "", "", true, false)

	assertContains(t, got, "--- file.txt")
	assertContains(t, got, "+++ /dev/null")
	assertContains(t, got, "@@")
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Fatalf("diff does not contain %q:\n%s", want, got)
	}
}
