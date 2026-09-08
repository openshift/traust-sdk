package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

func formatAndWrite(path string, src []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("generated invalid Go for %s: %w", filepath.Base(path), err)
	}
	return os.WriteFile(path, formatted, 0o644)
}
