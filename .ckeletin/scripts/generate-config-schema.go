// scripts/generate-config-schema.go
//
// Generates a JSON Schema from the config registry.
// This enables agents and tools to validate config.yaml files programmatically
// without reading Go source code.
//
// Usage: go run scripts/generate-config-schema.go

//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	_ "github.com/peiman/ckeletin-go/.ckeletin/pkg/config/commands" // Import to trigger init() functions (framework)
	_ "github.com/peiman/ckeletin-go/internal/config/commands"      // Import to trigger init() functions (project)
)

// JSONSchema represents a JSON Schema document.
type JSONSchema struct {
	Schema      string                 `json:"$schema"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`

	// Property-level fields
	PropertyType string      `json:"type,omitempty"`
	Default      interface{} `json:"default,omitempty"`
	Enum         []string    `json:"enum,omitempty"`
	Items        *JSONSchema `json:"items,omitempty"`
	Desc         string      `json:"description,omitempty"`
	Examples     []string    `json:"examples,omitempty"`
}

func main() {
	options := config.Registry()

	schema := &schemaBuilder{
		root: map[string]interface{}{
			"$schema":     "https://json-schema.org/draft/2020-12/schema",
			"title":       "ckeletin-go configuration",
			"description": "Configuration schema for ckeletin-go CLI applications. Generated from the config registry.",
			"type":        "object",
			"properties":  map[string]interface{}{},
		},
	}

	var required []string

	for _, opt := range options {
		schema.addOption(opt)
		if opt.Required {
			required = append(required, opt.Key)
		}
	}

	if len(required) > 0 {
		schema.root["required"] = required
	}

	encoded, err := json.MarshalIndent(schema.root, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal schema: %v\n", err)
		os.Exit(1)
	}

	outFile := "config.schema.json"
	if err := os.WriteFile(outFile, append(encoded, '\n'), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write schema file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated config schema with %d options in %s\n", len(options), outFile)
}

type schemaBuilder struct {
	root map[string]interface{}
}

// addOption adds a config option to the schema, nesting by dot-separated path.
// e.g., "app.log.file_enabled" becomes { app: { properties: { log: { properties: { file_enabled: ... } } } } }
func (s *schemaBuilder) addOption(opt config.ConfigOption) {
	parts := strings.Split(opt.Key, ".")
	current := s.root["properties"].(map[string]interface{})

	// Navigate/create nested objects for all but the last part
	for _, part := range parts[:len(parts)-1] {
		if existing, ok := current[part]; ok {
			obj := existing.(map[string]interface{})
			if props, ok := obj["properties"]; ok {
				current = props.(map[string]interface{})
			} else {
				props := map[string]interface{}{}
				obj["properties"] = props
				current = props
			}
		} else {
			props := map[string]interface{}{}
			current[part] = map[string]interface{}{
				"type":       "object",
				"properties": props,
			}
			current = props
		}
	}

	// Add the leaf property
	leaf := parts[len(parts)-1]
	prop := map[string]interface{}{
		"type": goTypeToJSONSchemaType(opt.Type),
	}

	if opt.Description != "" {
		prop["description"] = opt.Description
	}

	if opt.DefaultValue != nil {
		prop["default"] = opt.DefaultValue
	}

	if opt.Example != "" {
		prop["examples"] = []string{opt.Example}
	}

	// For string arrays
	if opt.Type == "[]string" || opt.Type == "stringslice" {
		prop["type"] = "array"
		prop["items"] = map[string]interface{}{"type": "string"}
	}

	current[leaf] = prop
}

func goTypeToJSONSchemaType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "bool":
		return "boolean"
	case "int":
		return "integer"
	case "float", "float64":
		return "number"
	case "[]string", "stringslice":
		return "array"
	default:
		return "string"
	}
}
