package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type structDef struct {
	name string
	node *SchemaNode
}

type typeRegistry struct {
	owner     map[string]string // structName -> schema filename
	defNames  map[string]string // filename#$defs/key -> structName
	inlineKey map[string]string // filename#inline#parent#field -> structName
}

func newTypeRegistry() *typeRegistry {
	return &typeRegistry{
		owner:     make(map[string]string),
		defNames:  make(map[string]string),
		inlineKey: make(map[string]string),
	}
}

func (r *typeRegistry) build(schemas map[string]*SchemaFile) {
	for _, filename := range sortedSchemaFilenames(schemas) {
		sf := schemas[filename]
		topLevel := schemaStructName(filename)
		r.owner[topLevel] = filename

		for _, defKey := range sortedStringKeysMapKeys(sf.Defs) {
			node := sf.Defs[defKey]
			if !shouldGenerateDefStruct(node) {
				continue
			}
			r.allocateDef(filename, topLevel, defKey)
		}
	}
}

func (r *typeRegistry) allocateDef(filename, topLevel, defKey string) string {
	fk := filename + "#" + defKey
	if name, ok := r.defNames[fk]; ok {
		return name
	}
	base := toPascalCase(defKey)
	if base == topLevel {
		base = topLevel + base
	}
	name := r.uniqueName(filename, topLevel, base)
	r.defNames[fk] = name
	r.owner[name] = filename
	return name
}

func (r *typeRegistry) allocateInline(filename, topLevel, parentStruct, fieldName string) string {
	suffix := inlineSuffix(fieldName)
	ik := fmt.Sprintf("%s#inline#%s#%s", filename, parentStruct, fieldName)
	if name, ok := r.inlineKey[ik]; ok {
		return name
	}
	base := parentStruct + suffix
	name := r.uniqueName(filename, topLevel, base)
	r.inlineKey[ik] = name
	r.owner[name] = filename
	return name
}

func (r *typeRegistry) uniqueName(filename, topLevel, base string) string {
	name := base
	if owner, exists := r.owner[name]; exists && owner != filename {
		name = topLevel + base
	}
	if owner, exists := r.owner[name]; exists && owner != filename {
		name = topLevel + toPascalCase(strings.TrimSuffix(filename, ".schema.json")) + base
	}
	return name
}

func (r *typeRegistry) defStructName(filename, defKey string) string {
	fk := filename + "#" + defKey
	if name, ok := r.defNames[fk]; ok {
		return name
	}
	return toPascalCase(defKey)
}

func (r *typeRegistry) structNameForRef(ref, sourceFile string) string {
	targetFile := refTargetFile(ref, sourceFile)
	defKey := refDefKey(ref)
	fk := targetFile + "#" + defKey
	if name, ok := r.defNames[fk]; ok {
		return name
	}
	return toPascalCase(defKey)
}

func inlineSuffix(fieldName string) string {
	if strings.HasSuffix(fieldName, "_item") {
		return toPascalCase(strings.TrimSuffix(fieldName, "_item")) + "Item"
	}
	return toPascalCase(fieldName)
}

func refTargetFile(ref, sourceFile string) string {
	hash := strings.Index(ref, "#")
	if hash < 0 {
		return sourceFile
	}
	filePart := ref[:hash]
	if filePart == "" {
		return sourceFile
	}
	return filePart
}

type fileGen struct {
	sf           *SchemaFile
	topLevelName string
	enumMappings []EnumMapping
	enumsPkg     string
	version      string
	registry     *typeRegistry
	structs      map[string]*structDef
	usesEnums    bool
}

func GenerateTypes(schemas map[string]*SchemaFile, enumMappings []EnumMapping, enumsPkg, outDir, contractsVersion string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outDir, err)
	}

	registry := newTypeRegistry()
	registry.build(schemas)

	filenames := sortedSchemaFilenames(schemas)
	for _, filename := range filenames {
		sf := schemas[filename]
		fg := newFileGen(sf, enumMappings, enumsPkg, contractsVersion, registry)
		if err := fg.collectStructs(); err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
		content, err := fg.emit()
		if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
		outPath := filepath.Join(outDir, schemaOutputFilename(filename))
		if err := formatAndWrite(outPath, []byte(content)); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
	}
	return nil
}

func sortedSchemaFilenames(schemas map[string]*SchemaFile) []string {
	filenames := make([]string, 0, len(schemas))
	for name := range schemas {
		filenames = append(filenames, name)
	}
	sort.Strings(filenames)
	return filenames
}

func schemaOutputFilename(filename string) string {
	base := strings.TrimSuffix(filename, ".schema.json")
	return base + ".go"
}

func schemaStructName(filename string) string {
	base := strings.TrimSuffix(filename, ".schema.json")
	return toPascalCase(base)
}

func newFileGen(sf *SchemaFile, enumMappings []EnumMapping, enumsPkg, version string, registry *typeRegistry) *fileGen {
	return &fileGen{
		sf:           sf,
		topLevelName: schemaStructName(sf.Filename),
		enumMappings: enumMappings,
		enumsPkg:     enumsPkg,
		version:      version,
		registry:     registry,
		structs:      make(map[string]*structDef),
	}
}

func (fg *fileGen) collectStructs() error {
	fg.registerStruct(fg.topLevelName, &SchemaNode{
		Properties:           fg.sf.Properties,
		Required:             fg.sf.Required,
		AdditionalProperties: fg.sf.AdditionalProperties,
		Type:                 "object",
	})

	defKeys := sortedStringKeysMapKeys(fg.sf.Defs)
	for _, defKey := range defKeys {
		node := fg.sf.Defs[defKey]
		if !shouldGenerateDefStruct(node) {
			continue
		}
		name := fg.registry.defStructName(fg.sf.Filename, defKey)
		fg.registerStruct(name, node)
	}
	return nil
}

func shouldGenerateDefStruct(node *SchemaNode) bool {
	if node == nil {
		return false
	}
	if len(node.Enum) > 0 && len(node.Properties) == 0 {
		return false
	}
	return len(node.Properties) > 0
}

func (fg *fileGen) registerStruct(name string, node *SchemaNode) {
	if node == nil || len(node.Properties) == 0 {
		return
	}
	if fg.registry.owner[name] != fg.sf.Filename {
		return
	}
	if _, exists := fg.structs[name]; exists {
		return
	}
	fg.structs[name] = &structDef{name: name, node: node}

	propNames := sortedStringKeysMapKeys(node.Properties)
	for _, propName := range propNames {
		propNode := node.Properties[propName]
		fg.registerNestedStructs(name, propName, propNode)
	}
}

func (fg *fileGen) registerNestedStructs(parentStruct, fieldName string, node *SchemaNode) {
	if node == nil {
		return
	}
	if node.Ref != "" && node.ResolvedRef != nil {
		return
	}
	if len(node.Enum) > 0 && len(node.Properties) == 0 && node.AdditionalProperties == nil {
		return
	}

	kind := fg.nodeKind(node)
	if len(node.Properties) > 0 {
		inlineName := fg.inlineStructName(parentStruct, fieldName)
		fg.registerStruct(inlineName, node)
	} else if node.AdditionalProperties != nil && (kind == "object" || kind == "") {
		entryName := fg.inlineStructName(parentStruct, fieldName+"_entry")
		if len(node.AdditionalProperties.Properties) > 0 {
			fg.registerStruct(entryName, node.AdditionalProperties)
		}
		fg.registerNestedStructs(parentStruct, fieldName+"_entry", node.AdditionalProperties)
	}

	if kind == "array" && node.Items != nil {
		fg.registerNestedStructs(parentStruct, fieldName+"_item", node.Items)
	}
}

func (fg *fileGen) inlineStructName(parentStruct, fieldName string) string {
	return fg.registry.allocateInline(fg.sf.Filename, fg.topLevelName, parentStruct, fieldName)
}

func (fg *fileGen) emit() (string, error) {
	var body strings.Builder

	structNames := make([]string, 0, len(fg.structs))
	for name := range fg.structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)

	if _, ok := fg.structs[fg.topLevelName]; ok {
		if err := fg.emitStruct(&body, fg.structs[fg.topLevelName]); err != nil {
			return "", err
		}
	}
	for _, name := range structNames {
		if name == fg.topLevelName {
			continue
		}
		if err := fg.emitStruct(&body, fg.structs[name]); err != nil {
			return "", err
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "// Code generated from traust-contracts v%s. DO NOT EDIT.\n\n", fg.version)
	fmt.Fprint(&b, "package types\n\n")

	if fg.usesEnums {
		fmt.Fprint(&b, "import (\n")
		fmt.Fprintf(&b, "\t%q\n", fg.enumsPkg)
		fmt.Fprint(&b, ")\n\n")
	}

	b.WriteString(body.String())
	return b.String(), nil
}

func (fg *fileGen) emitStruct(b *strings.Builder, def *structDef) error {
	requiredSet := stringSet(def.node.Required)
	propNames := sortedStringKeysMapKeys(def.node.Properties)

	fmt.Fprintf(b, "type %s struct {\n", def.name)
	for _, propName := range propNames {
		propNode := def.node.Properties[propName]
		required := isRequired(requiredSet, propName)
		goType, err := fg.goType(propNode, required, def.name, propName)
		if err != nil {
			return fmt.Errorf("property %q on %s: %w", propName, def.name, err)
		}
		fieldName := toPascalCase(propName)
		tag := jsonTag(propName, required)
		fmt.Fprintf(b, "\t%s %s `%s`\n", fieldName, goType, tag)
	}
	fmt.Fprint(b, "}\n\n")
	return nil
}

func jsonTag(name string, required bool) string {
	if required {
		return fmt.Sprintf("json:%q", name)
	}
	return fmt.Sprintf("json:%q", name+",omitempty")
}

func (fg *fileGen) goType(node *SchemaNode, required bool, parentStruct, fieldName string) (string, error) {
	if node == nil {
		return wrapOptional("interface{}", required), nil
	}

	if node.Ref != "" {
		if node.ResolvedRef == nil {
			return wrapOptional("interface{}", required), nil
		}
		base, err := fg.goTypeForResolved(node.ResolvedRef, node.Ref, parentStruct, fieldName)
		if err != nil {
			return "", err
		}
		return wrapOptional(base, required), nil
	}

	base, err := fg.goTypeBase(node, parentStruct, fieldName)
	if err != nil {
		return "", err
	}
	return wrapOptional(base, required), nil
}

func (fg *fileGen) goTypeForResolved(node *SchemaNode, ref, parentStruct, fieldName string) (string, error) {
	if isEnumOnlyDef(node) {
		return fg.enumOrString(node.Enum), nil
	}
	if len(node.Properties) > 0 {
		return fg.registry.structNameForRef(ref, fg.sf.Filename), nil
	}
	if node.AdditionalProperties != nil {
		return fg.mapType(node.AdditionalProperties, parentStruct, fieldName), nil
	}
	return "map[string]interface{}", nil
}

func (fg *fileGen) goTypeBase(node *SchemaNode, parentStruct, fieldName string) (string, error) {
	kind := fg.nodeKind(node)

	if len(node.Enum) > 0 && kind != "integer" && kind != "number" {
		return fg.enumOrString(node.Enum), nil
	}

	switch kind {
	case "string":
		return "string", nil
	case "integer":
		return "int", nil
	case "number":
		return "float64", nil
	case "boolean":
		return "bool", nil
	case "union":
		return "interface{}", nil
	case "array":
		if node.Items == nil {
			return "[]interface{}", nil
		}
		elem, err := fg.goType(node.Items, true, parentStruct, fieldName+"_item")
		if err != nil {
			return "", err
		}
		return "[]" + unwrapPointer(elem), nil
	case "object":
		if len(node.Properties) > 0 {
			return fg.inlineStructName(parentStruct, fieldName), nil
		}
		if node.AdditionalProperties != nil {
			return fg.mapType(node.AdditionalProperties, parentStruct, fieldName), nil
		}
		return "map[string]interface{}", nil
	default:
		if len(node.Properties) > 0 {
			return fg.inlineStructName(parentStruct, fieldName), nil
		}
		return "interface{}", nil
	}
}

func (fg *fileGen) nodeKind(node *SchemaNode) string {
	if node.Type == nil {
		if len(node.Properties) > 0 {
			return "object"
		}
		return ""
	}
	switch t := node.Type.(type) {
	case string:
		return t
	case []string:
		return "union"
	default:
		return ""
	}
}

func (fg *fileGen) enumOrString(values []string) string {
	if mapping := MatchEnum(values, fg.enumMappings); mapping != nil {
		fg.usesEnums = true
		return "enums." + mapping.GoType
	}
	return "string"
}

func refDefKey(ref string) string {
	hash := strings.Index(ref, "#")
	if hash < 0 {
		return ref
	}
	fragment := strings.TrimPrefix(ref[hash+1:], "/")
	parts := strings.Split(fragment, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return fragment
}

func (fg *fileGen) mapType(valueNode *SchemaNode, parentStruct, fieldName string) string {
	if valueNode == nil || len(valueNode.Properties) == 0 && valueNode.Ref == "" && valueNode.Type == nil {
		return "map[string]interface{}"
	}
	entryField := fieldName + "_entry"
	elem, err := fg.goType(valueNode, true, parentStruct, entryField)
	if err != nil {
		return "map[string]interface{}"
	}
	return "map[string]" + unwrapPointer(elem)
}

func isEnumOnlyDef(node *SchemaNode) bool {
	return len(node.Enum) > 0 && len(node.Properties) == 0
}

func wrapOptional(goType string, required bool) string {
	if required {
		return goType
	}
	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[") || goType == "interface{}" {
		return goType
	}
	return "*" + goType
}

func unwrapPointer(goType string) string {
	if strings.HasPrefix(goType, "*") {
		return goType[1:]
	}
	return goType
}

func sortedStringKeysMapKeys(m map[string]*SchemaNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func isRequired(requiredSet map[string]struct{}, name string) bool {
	_, ok := requiredSet[name]
	return ok
}
