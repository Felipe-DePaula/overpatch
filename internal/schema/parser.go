package schema

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParseFile reads an Overpatch document from the given file path, unmarshals
// it, and validates it. It returns the parsed document or the first error
// encountered.
func ParseFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if err := ValidateDocument(&doc); err != nil {
		return nil, err
	}

	return &doc, nil
}
