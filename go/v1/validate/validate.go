// Package validate checks JSON bytes against the v1 contract schemas embedded from
// schemas/v1 at generation time. Look up a schema by its basename minus ".schema.json"
// (e.g. "report", "layer", "triage").
package validate

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaBaseURL = "https://example.com/traust-contracts/schemas/v1/"

// CompiledSchema holds a pre-compiled schema for fast repeated validation.
type CompiledSchema struct {
	schema *jsonschema.Schema
}

// Validate checks data against this pre-compiled schema.
func (c *CompiledSchema) Validate(data []byte) error {
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}
	if err := c.schema.Validate(instance); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// Validate checks raw JSON data against the named schema.
// Schema name is the filename without .schema.json (e.g., "report", "layer", "triage").
func Validate(schema string, data []byte) error {
	return ValidateBytes(schema, data)
}

// ValidateBytes is an alias for Validate (clarity for callers with []byte).
func ValidateBytes(schema string, data []byte) error {
	compiled, err := compileSchema(schema)
	if err != nil {
		return err
	}
	return compiled.Validate(data)
}

// SchemaNames returns all available schema names.
func SchemaNames() []string {
	schemaNamesOnce.Do(func() {
		names := make([]string, 0, len(Schemas))
		for name := range Schemas {
			names = append(names, name)
		}
		sort.Strings(names)
		schemaNames = names
	})
	return append([]string(nil), schemaNames...)
}

// MustCompile pre-compiles a schema for repeated validation. Panics on unknown schema.
func MustCompile(schema string) *CompiledSchema {
	compiled, err := compileSchema(schema)
	if err != nil {
		panic(err)
	}
	return compiled
}

type mapLoader struct{}

func (mapLoader) Load(rawURL string) (any, error) {
	filename, err := schemaFilename(rawURL)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSuffix(filename, ".schema.json")
	content, ok := Schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema %q not found in generated map", name)
	}
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse schema %q: %w", name, err)
	}
	return doc, nil
}

func schemaFilename(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse schema URL %q: %w", rawURL, err)
	}

	switch {
	case strings.HasPrefix(rawURL, schemaBaseURL):
		return path.Base(u.Path), nil
	case u.Scheme == "file":
		name := path.Base(u.Path)
		if strings.HasSuffix(name, ".schema.json") {
			return name, nil
		}
	case strings.HasSuffix(rawURL, ".schema.json"):
		return path.Base(u.Path), nil
	}

	return "", fmt.Errorf("unknown schema URL %q", rawURL)
}

func schemaResourceURL(name string) (string, error) {
	if !schemaExists(name) {
		return "", fmt.Errorf("unknown schema %q", name)
	}
	return schemaBaseURL + name + ".schema.json", nil
}

func schemaExists(name string) bool {
	_, ok := Schemas[name]
	return ok
}

var (
	schemaNames     []string
	schemaNamesOnce sync.Once

	compilerOnce sync.Once
	compiler     *jsonschema.Compiler

	compiledCache sync.Map
)

func getCompiler() *jsonschema.Compiler {
	compilerOnce.Do(func() {
		c := jsonschema.NewCompiler()
		c.UseLoader(mapLoader{})
		compiler = c
	})
	return compiler
}

func compileSchema(name string) (*CompiledSchema, error) {
	if cached, ok := compiledCache.Load(name); ok {
		return cached.(*CompiledSchema), nil
	}

	resourceURL, err := schemaResourceURL(name)
	if err != nil {
		return nil, err
	}

	sch, err := getCompiler().Compile(resourceURL)
	if err != nil {
		return nil, fmt.Errorf("compile schema %q: %w", name, err)
	}

	compiled := &CompiledSchema{schema: sch}
	actual, _ := compiledCache.LoadOrStore(name, compiled)
	return actual.(*CompiledSchema), nil
}
