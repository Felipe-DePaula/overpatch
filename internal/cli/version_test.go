package cli

import (
	"bytes"
	"testing"
)

func TestVersionOutputGoesToBuffer(t *testing.T) {
	oldVersion := Version
	Version = "v0.1.0-test"
	t.Cleanup(func() {
		Version = oldVersion
	})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"version"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if got, want := buf.String(), "overpatch v0.1.0-test\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
