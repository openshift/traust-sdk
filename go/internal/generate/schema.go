package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SchemaFile is a parsed JSON Schema document with a resolved type tree.
type SchemaFile struct {
	ID                   string
	Title                string
	Filename             string
	Properties           map[string]*SchemaNode
	Required             []string
	Defs                 map[string]*SchemaNode
	AdditionalProperties *SchemaNode
}

// SchemaNode is a node in a schema type tree.
type SchemaNode struct {
	Type                 interface{} // string or []string for union types
	Ref                  string
	ResolvedRef          *SchemaNode
	RefSource            string
	Properties           map[string]*SchemaNode
	Required             []string
	Items                *SchemaNode
	Enum                 []string
	Format               string
	Description          string
	AdditionalProperties *SchemaNode
	MinItems             *int
	Const                interface{}
}

type rawSchemaFile struct {
	ID                   string                     `json:"$id"`
	Title                string                     `json:"title"`
	Required             []string                   `json:"required"`
	Properties           map[string]json.RawMessage `json:"properties"`
	Defs                 map[string]json.RawMessage `json:"$defs"`
	Definitions          map[string]json.RawMessage `json:"definitions"`
	AdditionalProperties json.RawMessage            `json:"additionalProperties"`
}

type rawSchemaNode struct {
	Type                 json.RawMessage            `json:"type"`
	Ref                  string                     `json:"$ref"`
	Properties           map[string]json.RawMessage `json:"properties"`
	Required             []string                   `json:"required"`
	Items                json.RawMessage            `json:"items"`
	Enum                 json.RawMessage            `json:"enum"`
	Format               string                     `json:"format"`
	Description          string                     `json:"description"`
	AdditionalProperties json.RawMessage            `json:"additionalProperties"`
	MinItems             *int                       `json:"minItems"`
	Const                json.RawMessage            `json:"const"`
}

// LoadSchemas loads all .schema.json files from dir into a filename-keyed map.
func LoadSchemas(dir string) (map[string]*SchemaFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading schemas directory %s: %w", dir, err)
	}

	schemas := make(map[string]*SchemaFile)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		var raw rawSchemaFile
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		sf := &SchemaFile{
			ID:       raw.ID,
			Title:    raw.Title,
			Filename: entry.Name(),
			Required: raw.Required,
			Defs:     make(map[string]*SchemaNode),
		}

		for name, propRaw := range raw.Properties {
			node, err := parseSchemaNode(propRaw)
			if err != nil {
				return nil, fmt.Errorf("parsing property %q in %s: %w", name, entry.Name(), err)
			}
			if sf.Properties == nil {
				sf.Properties = make(map[string]*SchemaNode)
			}
			sf.Properties[name] = node
		}

		for name, defRaw := range raw.Defs {
			node, err := parseSchemaNode(defRaw)
			if err != nil {
				return nil, fmt.Errorf("parsing $defs/%q in %s: %w", name, entry.Name(), err)
			}
			sf.Defs[name] = node
		}
		for name, defRaw := range raw.Definitions {
			if _, exists := sf.Defs[name]; exists {
				continue
			}
			node, err := parseSchemaNode(defRaw)
			if err != nil {
				return nil, fmt.Errorf("parsing definitions/%q in %s: %w", name, entry.Name(), err)
			}
			sf.Defs[name] = node
		}

		ap, err := parseAdditionalProperties(raw.AdditionalProperties)
		if err != nil {
			return nil, fmt.Errorf("parsing additionalProperties in %s: %w", entry.Name(), err)
		}
		sf.AdditionalProperties = ap

		schemas[entry.Name()] = sf
	}

	return schemas, nil
}

// ResolveRefs resolves every $ref pointer in the loaded schemas.
func ResolveRefs(schemas map[string]*SchemaFile) error {
	visiting := make(map[*SchemaNode]bool)
	for _, sf := range schemas {
		if err := resolveSchemaFileRefs(sf, schemas, visiting); err != nil {
			return err
		}
	}
	return nil
}

func resolveSchemaFileRefs(sf *SchemaFile, schemas map[string]*SchemaFile, visiting map[*SchemaNode]bool) error {
	for _, node := range sf.Properties {
		if err := resolveNodeRefs(node, sf.Filename, schemas, visiting); err != nil {
			return fmt.Errorf("%s: %w", sf.Filename, err)
		}
	}
	for name, node := range sf.Defs {
		if err := resolveNodeRefs(node, sf.Filename, schemas, visiting); err != nil {
			return fmt.Errorf("%s/$defs/%s: %w", sf.Filename, name, err)
		}
	}
	if sf.AdditionalProperties != nil {
		if err := resolveNodeRefs(sf.AdditionalProperties, sf.Filename, schemas, visiting); err != nil {
			return fmt.Errorf("%s additionalProperties: %w", sf.Filename, err)
		}
	}
	return nil
}

func resolveNodeRefs(node *SchemaNode, sourceFile string, schemas map[string]*SchemaFile, visiting map[*SchemaNode]bool) error {
	if node == nil {
		return nil
	}
	if visiting[node] {
		return nil
	}
	visiting[node] = true
	defer delete(visiting, node)

	for _, prop := range node.Properties {
		if err := resolveNodeRefs(prop, sourceFile, schemas, visiting); err != nil {
			return err
		}
	}
	if node.Items != nil {
		if err := resolveNodeRefs(node.Items, sourceFile, schemas, visiting); err != nil {
			return err
		}
	}
	if node.AdditionalProperties != nil {
		if err := resolveNodeRefs(node.AdditionalProperties, sourceFile, schemas, visiting); err != nil {
			return err
		}
	}

	if node.Ref == "" {
		return nil
	}

	target, refSource, err := lookupRef(node.Ref, sourceFile, schemas)
	if err != nil {
		return err
	}
	if err := resolveNodeRefs(target, refSource, schemas, visiting); err != nil {
		return err
	}
	node.ResolvedRef = target
	node.RefSource = refSource
	return nil
}

func lookupRef(ref, sourceFile string, schemas map[string]*SchemaFile) (*SchemaNode, string, error) {
	hash := strings.Index(ref, "#")
	if hash < 0 {
		return nil, "", fmt.Errorf("invalid $ref %q: missing fragment", ref)
	}

	filePart := ref[:hash]
	fragment := strings.TrimPrefix(ref[hash+1:], "/")

	targetFile := sourceFile
	if filePart != "" {
		targetFile = filePart
	}

	sf, ok := schemas[targetFile]
	if !ok {
		return nil, "", fmt.Errorf("schema file %q not found for $ref %q", targetFile, ref)
	}

	parts := strings.Split(fragment, "/")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("unsupported $ref fragment %q in %q", fragment, ref)
	}

	switch parts[0] {
	case "$defs", "definitions":
	default:
		return nil, "", fmt.Errorf("unsupported $ref path %q in %q", parts[0], ref)
	}

	defName := parts[1]
	node, ok := sf.Defs[defName]
	if !ok {
		return nil, "", fmt.Errorf("definition %q not found in %s", defName, targetFile)
	}

	return node, targetFile, nil
}

func parseSchemaNode(raw json.RawMessage) (*SchemaNode, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var rn rawSchemaNode
	if err := json.Unmarshal(raw, &rn); err != nil {
		return nil, err
	}

	enum, err := parseEnumField(rn.Enum)
	if err != nil {
		return nil, fmt.Errorf("enum: %w", err)
	}

	node := &SchemaNode{
		Ref:         rn.Ref,
		Required:    rn.Required,
		Enum:        enum,
		Format:      rn.Format,
		Description: rn.Description,
		MinItems:    rn.MinItems,
	}

	node.Type = parseTypeField(rn.Type)
	node.Const = parseConstField(rn.Const)

	for name, propRaw := range rn.Properties {
		prop, err := parseSchemaNode(propRaw)
		if err != nil {
			return nil, fmt.Errorf("property %q: %w", name, err)
		}
		if node.Properties == nil {
			node.Properties = make(map[string]*SchemaNode)
		}
		node.Properties[name] = prop
	}

	if len(rn.Items) > 0 {
		items, err := parseSchemaNode(rn.Items)
		if err != nil {
			return nil, fmt.Errorf("items: %w", err)
		}
		node.Items = items
	}

	ap, err := parseAdditionalProperties(rn.AdditionalProperties)
	if err != nil {
		return nil, fmt.Errorf("additionalProperties: %w", err)
	}
	node.AdditionalProperties = ap

	return node, nil
}

func parseTypeField(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return single
	}

	var union []string
	if err := json.Unmarshal(raw, &union); err == nil {
		return union
	}

	return nil
}

func parseConstField(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}

	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return v
}

func parseEnumField(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var values []interface{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}

	enum := make([]string, 0, len(values))
	for _, v := range values {
		switch typed := v.(type) {
		case string:
			enum = append(enum, typed)
		case float64:
			enum = append(enum, fmt.Sprintf("%g", typed))
		case bool:
			enum = append(enum, fmt.Sprintf("%t", typed))
		default:
			enum = append(enum, fmt.Sprint(v))
		}
	}
	return enum, nil
}

func parseAdditionalProperties(raw json.RawMessage) (*SchemaNode, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var allowed bool
	if err := json.Unmarshal(raw, &allowed); err == nil {
		if !allowed {
			return nil, nil
		}
		return &SchemaNode{}, nil
	}

	return parseSchemaNode(raw)
}
