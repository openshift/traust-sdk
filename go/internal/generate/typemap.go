package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnumMapping links a schema enum value set to a generated Go enum type.
type EnumMapping struct {
	JSONName string
	GoType   string
	Values   []string
}

// LoadEnumMappings loads all enums/*.json files into enum mappings.
func LoadEnumMappings(enumsDir string) ([]EnumMapping, error) {
	entries, err := os.ReadDir(enumsDir)
	if err != nil {
		return nil, fmt.Errorf("reading enums directory %s: %w", enumsDir, err)
	}

	var mappings []EnumMapping
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(enumsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		var def enumDef
		if err := json.Unmarshal(data, &def); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		if def.Name == "" || len(def.Values) == 0 {
			continue
		}

		mappings = append(mappings, EnumMapping{
			JSONName: def.Name,
			GoType:   toPascalCase(def.Name),
			Values:   append([]string(nil), def.Values...),
		})
	}

	return mappings, nil
}

// MatchEnum returns the best enum mapping for schema values, or nil if none match.
// Exact set matches are preferred over subset matches; among subset matches the
// smallest superset mapping wins.
func MatchEnum(values []string, mappings []EnumMapping) *EnumMapping {
	if len(values) == 0 {
		return nil
	}

	schemaSet := stringSet(values)

	var bestSubset *EnumMapping
	bestSubsetSize := int(^uint(0) >> 1)

	for i := range mappings {
		mapping := &mappings[i]
		mappingSet := stringSet(mapping.Values)

		if setsEqual(schemaSet, mappingSet) {
			return mapping
		}

		if isSubset(schemaSet, mappingSet) && len(mappingSet) < bestSubsetSize {
			bestSubset = mapping
			bestSubsetSize = len(mappingSet)
		}
	}

	return bestSubset
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return set
}

func setsEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func isSubset(sub, sup map[string]struct{}) bool {
	for k := range sub {
		if _, ok := sup[k]; !ok {
			return false
		}
	}
	return true
}
