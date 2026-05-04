package safety

import (
	"strings"
	"testing"
)

func TestValidateOperationPath(t *testing.T) {
	tests := []struct {
		path        string
		expectError string
	}{
		// Valid cases
		{path: "README.md", expectError: ""},
		{path: "docs/PROTOCOL.md", expectError: ""},
		{path: "src/auth.ts", expectError: ""},
		{path: "internal/schema/validate.go", expectError: ""},
		{path: ".gitignore", expectError: ""},
		{path: "docs/environment.md", expectError: ""},
		{path: "src/env.ts", expectError: ""},
		{path: "nested\\path\\file.txt", expectError: ""},

		// Invalid cases
		{path: "", expectError: "path is empty"},
		{path: "   ", expectError: "path is empty"},
		{path: "/etc/passwd", expectError: "path must be relative"},
		{path: "C:\\Users\\felip\\.env", expectError: "path must be relative"},
		{path: "D:/Dev/file.txt", expectError: "path must be relative"},
		{path: "\\\\server\\share\\file.txt", expectError: "path must be relative"},
		{path: "../README.md", expectError: "path traversal is not allowed"},
		{path: "docs/../../README.md", expectError: "path traversal is not allowed"},
		{path: "..\\README.md", expectError: "path traversal is not allowed"},
		{path: ".git/config", expectError: "path is blocked"},
		{path: "docs/.git/config", expectError: "path is blocked"},
		{path: ".env", expectError: "path is blocked"},
		{path: ".env.local", expectError: "path is blocked"},
		{path: "config/.env", expectError: "path is blocked"},
		{path: ".ssh/id_rsa", expectError: "path is blocked"},
		{path: "node_modules/pkg/index.js", expectError: "path is blocked"},
	}

	for _, tc := range tests {
		err := ValidateOperationPath(tc.path)
		if tc.expectError == "" {
			if err != nil {
				t.Errorf("expected path %q to be valid, got: %v", tc.path, err)
			}
		} else {
			if err == nil {
				t.Errorf("expected path %q to fail with %q, got nil", tc.path, tc.expectError)
			} else if !strings.Contains(err.Error(), tc.expectError) {
				t.Errorf("expected path %q to fail with %q, got: %v", tc.path, tc.expectError, err)
			}
		}
	}
}
