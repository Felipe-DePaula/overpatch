// Package diff renders textual previews of staged file changes.
package diff

import "strings"

// Unified renders a simple unified diff for a single file.
func Unified(path string, original string, staged string, originalExists bool, stagedExists bool) string {
	var builder strings.Builder

	builder.WriteString("--- ")
	if originalExists {
		builder.WriteString(path)
	} else {
		builder.WriteString("/dev/null")
	}
	builder.WriteByte('\n')

	builder.WriteString("+++ ")
	if stagedExists {
		builder.WriteString(path)
	} else {
		builder.WriteString("/dev/null")
	}
	builder.WriteByte('\n')

	builder.WriteString("@@\n")
	if originalExists {
		builder.WriteString(prefixLines("-", original))
	}
	if stagedExists {
		builder.WriteString(prefixLines("+", staged))
	}

	return builder.String()
}

func prefixLines(prefix string, content string) string {
	if content == "" {
		return ""
	}

	lines := strings.SplitAfter(content, "\n")
	var builder strings.Builder
	for _, line := range lines {
		if line == "" {
			continue
		}
		builder.WriteString(prefix)
		builder.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}
