package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GenerateSchemasEmbed(schemasDir, outFile, contractsVersion string) error {
	entries, err := os.ReadDir(schemasDir)
	if err != nil {
		return fmt.Errorf("read schemas dir: %w", err)
	}

	type schemaEntry struct {
		name    string
		content string
	}

	var schemas []schemaEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(schemasDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		name := strings.TrimSuffix(entry.Name(), ".schema.json")
		schemas = append(schemas, schemaEntry{name: name, content: string(data)})
	}

	sort.Slice(schemas, func(i, j int) bool { return schemas[i].name < schemas[j].name })

	var b strings.Builder
	fmt.Fprintf(&b, "// Code generated from traust-contracts %s. DO NOT EDIT.\n\n", contractsVersion)
	fmt.Fprint(&b, "package validate\n\n")
	fmt.Fprint(&b, "// Schemas maps schema name to raw JSON Schema content.\n")
	fmt.Fprint(&b, "var Schemas = map[string]string{\n")
	for _, s := range schemas {
		fmt.Fprintf(&b, "\t%q: %s,\n", s.name, goRawStringLiteral(s.content))
	}
	fmt.Fprint(&b, "}\n")

	return formatAndWrite(outFile, []byte(b.String()))
}

func goRawStringLiteral(s string) string {
	if !strings.Contains(s, "`") {
		return "`" + s + "`"
	}
	return fmt.Sprintf("%q", s)
}
