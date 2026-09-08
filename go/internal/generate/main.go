package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	enumsPkg = "github.com/openshift/traust-sdk/go/v1/enums"
)

func main() {
	enumsDir := flag.String("enums", "../../enums", "path to enums/ directory containing JSON files")
	outDir := flag.String("out", "./enums", "output directory for generated Go files")
	schemasDir := flag.String("schemas", "", "path to schemas/ directory (optional; enables type generation)")
	typesDir := flag.String("types", "", "output directory for generated type files")
	validateFile := flag.String("validate", "", "output path for generated schemas_gen.go (e.g. ./validate/schemas_gen.go)")
	flag.Parse()

	entries, err := os.ReadDir(*enumsDir)
	if err != nil {
		log.Fatalf("reading enums directory %s: %v", *enumsDir, err)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("creating output directory %s: %v", *outDir, err)
	}

	var generated []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		src := filepath.Join(*enumsDir, entry.Name())
		if err := generateEnum(src, *outDir); err != nil {
			log.Fatalf("generating from %s: %v", entry.Name(), err)
		}
		generated = append(generated, entry.Name())
	}

	fmt.Printf("generated %d enum files in %s\n", len(generated), *outDir)

	if *schemasDir != "" && *typesDir != "" {
		schemas, err := LoadSchemas(*schemasDir)
		if err != nil {
			log.Fatalf("loading schemas: %v", err)
		}
		if err := ResolveRefs(schemas); err != nil {
			log.Fatalf("resolving schema refs: %v", err)
		}
		enumMappings, err := LoadEnumMappings(*enumsDir)
		if err != nil {
			log.Fatalf("loading enum mappings: %v", err)
		}
		contractsVersion := readVersion(filepath.Join(*schemasDir, "..", "..", "VERSION"))
		if err := GenerateTypes(schemas, enumMappings, enumsPkg, *typesDir, contractsVersion); err != nil {
			log.Fatalf("generating types: %v", err)
		}
		if err := generateVersionFile(*typesDir, contractsVersion); err != nil {
			log.Fatalf("generating version.go: %v", err)
		}
		fmt.Printf("generated %d type files in %s\n", len(schemas), *typesDir)
	}

	if *schemasDir != "" && *validateFile != "" {
		contractsVersion := readVersion(filepath.Join(*schemasDir, "..", "..", "VERSION"))
		if err := GenerateSchemasEmbed(*schemasDir, *validateFile, contractsVersion); err != nil {
			log.Fatalf("generating schemas embed: %v", err)
		}
		fmt.Printf("generated schemas map in %s\n", *validateFile)
	}
}

func generateVersionFile(typesDir, contractsVersion string) error {
	content := fmt.Sprintf(
		"// Code generated from traust-contracts v%s. DO NOT EDIT.\n\npackage types\n\nconst ContractsVersion = %q\n",
		contractsVersion, contractsVersion,
	)
	path := filepath.Join(typesDir, "version.go")
	return formatAndWrite(path, []byte(content))
}

func readVersion(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}
