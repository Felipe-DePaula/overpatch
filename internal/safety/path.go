package safety

import (
	"fmt"
	"path"
	"strings"
)

// ValidateOperationPath performs lexical validation on a file path to ensure
// it is safe to operate on. It rejects absolute paths, path traversal,
// and known sensitive files or directories without accessing the filesystem.
func ValidateOperationPath(p string) error {
	p = strings.TrimSpace(p)
	if p == "" {
		return fmt.Errorf("path is empty")
	}

	// Normalize Windows separators to Unix separators for lexical analysis
	normalized := strings.ReplaceAll(p, "\\", "/")

	// Reject absolute paths
	// Unix absolute: starts with / (also covers Windows UNC like \\server\...)
	// Windows absolute: starts with drive letter (e.g. C:)
	if strings.HasPrefix(normalized, "/") {
		return fmt.Errorf("path must be relative")
	}
	if len(normalized) >= 2 && normalized[1] == ':' {
		return fmt.Errorf("path must be relative")
	}

	// Clean the path to resolve . and ..
	cleaned := path.Clean(normalized)

	// Reject path traversal
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("path traversal is not allowed")
	}

	// Check for blocked components
	for _, part := range strings.Split(normalized, "/") {
		if part == ".git" || part == ".ssh" || part == "node_modules" {
			return fmt.Errorf("path is blocked")
		}
		if part == ".env" || strings.HasPrefix(part, ".env.") {
			return fmt.Errorf("path is blocked")
		}
	}

	return nil
}
