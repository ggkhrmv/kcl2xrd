package generator

import (
	"strings"
	"testing"

	"github.com/ggkhrmv/kcl2xrd/pkg/parser"
	"gopkg.in/yaml.v3"
)

func TestGenerateXRD(t *testing.T) {
	schema := &parser.Schema{
		Name:        "TestResource",
		Description: "A test resource",
		Fields: []parser.Field{
			{
				Name:     "field1",
				Type:     "str",
				Required: true,
			},
			{
				Name:     "field2",
				Type:     "int",
				Required: false,
				Default:  "42",
			},
		},
	}

	xrdYAML, err := GenerateXRD(schema, "example.org", "v1alpha1")
	if err != nil {
		t.Fatalf("GenerateXRD failed: %v", err)
	}

	// Check that it's valid YAML
	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Generated XRD is not valid YAML: %v", err)
	}

	// Check basic structure
	if xrd["apiVersion"] != "apiextensions.crossplane.io/v1" {
		t.Errorf("Expected apiVersion 'apiextensions.crossplane.io/v1', got '%v'", xrd["apiVersion"])
	}

	if xrd["kind"] != "CompositeResourceDefinition" {
		t.Errorf("Expected kind 'CompositeResourceDefinition', got '%v'", xrd["kind"])
	}

	// Check metadata
	metadata, ok := xrd["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("metadata is not a map")
	}

	expectedName := "testresources.example.org"
	if metadata["name"] != expectedName {
		t.Errorf("Expected metadata.name '%s', got '%v'", expectedName, metadata["name"])
	}

	// Check spec
	spec, ok := xrd["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("spec is not a map")
	}

	if spec["group"] != "example.org" {
		t.Errorf("Expected spec.group 'example.org', got '%v'", spec["group"])
	}
}

func TestConvertFieldToPropertySchema(t *testing.T) {
	tests := []struct {
		name     string
		field    parser.Field
		expected string // expected type
	}{
		{
			name:     "string field",
			field:    parser.Field{Name: "test", Type: "str"},
			expected: "string",
		},
		{
			name:     "integer field",
			field:    parser.Field{Name: "test", Type: "int"},
			expected: "integer",
		},
		{
			name:     "boolean field",
			field:    parser.Field{Name: "test", Type: "bool"},
			expected: "boolean",
		},
		{
			name:     "float field",
			field:    parser.Field{Name: "test", Type: "float"},
			expected: "number",
		},
		{
			name:     "array field",
			field:    parser.Field{Name: "test", Type: "[str]"},
			expected: "array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := convertFieldToPropertySchema(tt.field)
			if schema.Type != tt.expected {
				t.Errorf("Expected type '%s', got '%s'", tt.expected, schema.Type)
			}
		})
	}
}

func TestGenerateXRDWithDefaults(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestResource",
		Fields: []parser.Field{
			{
				Name:     "stringField",
				Type:     "str",
				Required: false,
				Default:  `"test"`,
			},
			{
				Name:     "intField",
				Type:     "int",
				Required: false,
				Default:  "42",
			},
			{
				Name:     "boolField",
				Type:     "bool",
				Required: false,
				Default:  "True",
			},
		},
	}

	xrdYAML, err := GenerateXRD(schema, "example.org", "v1alpha1")
	if err != nil {
		t.Fatalf("GenerateXRD failed: %v", err)
	}

	// Check that defaults are present in output
	if !strings.Contains(xrdYAML, "default:") {
		t.Error("Expected XRD to contain default values")
	}

	// Check that boolean default is not quoted
	if strings.Contains(xrdYAML, `default: "true"`) || strings.Contains(xrdYAML, `default: "True"`) {
		t.Error("Boolean default should not be quoted")
	}
}

func TestGenerateXRDWithClaims(t *testing.T) {
	schema := &parser.Schema{
		Name: "XTestResource",
		Fields: []parser.Field{
			{
				Name:     "field1",
				Type:     "str",
				Required: true,
			},
		},
	}

	opts := XRDOptions{
		Group:      "example.org",
		Version:    "v1alpha1",
		WithClaims: true,
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, opts)
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	// Parse the YAML
	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Generated XRD is not valid YAML: %v", err)
	}

	// Check spec
	spec, ok := xrd["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("spec is not a map")
	}

	// Check claimNames exists
	claimNames, ok := spec["claimNames"].(map[string]interface{})
	if !ok {
		t.Fatal("claimNames is not present or not a map")
	}

	// Check that 'X' prefix was removed
	if claimNames["kind"] != "TestResource" {
		t.Errorf("Expected claim kind 'TestResource', got '%v'", claimNames["kind"])
	}

	if claimNames["plural"] != "testresources" {
		t.Errorf("Expected claim plural 'testresources', got '%v'", claimNames["plural"])
	}
}

func TestGenerateXRDWithCustomClaimNames(t *testing.T) {
	schema := &parser.Schema{
		Name: "XTestResource",
		Fields: []parser.Field{
			{
				Name:     "field1",
				Type:     "str",
				Required: true,
			},
		},
	}

	opts := XRDOptions{
		Group:       "example.org",
		Version:     "v1alpha1",
		WithClaims:  true,
		ClaimKind:   "CustomClaim",
		ClaimPlural: "customclaims",
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, opts)
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	// Parse the YAML
	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Generated XRD is not valid YAML: %v", err)
	}

	// Check spec
	spec, ok := xrd["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("spec is not a map")
	}

	// Check claimNames
	claimNames, ok := spec["claimNames"].(map[string]interface{})
	if !ok {
		t.Fatal("claimNames is not present or not a map")
	}

	// Check custom names are used
	if claimNames["kind"] != "CustomClaim" {
		t.Errorf("Expected claim kind 'CustomClaim', got '%v'", claimNames["kind"])
	}

	if claimNames["plural"] != "customclaims" {
		t.Errorf("Expected claim plural 'customclaims', got '%v'", claimNames["plural"])
	}
}

func TestConvertFieldWithAnyType(t *testing.T) {
	// Test that 'any' type fields don't get type: object
	field := parser.Field{
		Name:                  "principal",
		Type:                  "any",
		Required:              false,
		Description:           "The principals this statement applies to",
		PreserveUnknownFields: true,
	}

	schema := convertFieldToPropertySchema(field)

	// 'any' type should NOT have a type set
	if schema.Type != "" {
		t.Errorf("Expected type to be empty for 'any' type, got '%s'", schema.Type)
	}

	// But should have PreserveUnknownFields
	if schema.XKubernetesPreserveUnknownFields == nil || !*schema.XKubernetesPreserveUnknownFields {
		t.Error("Expected x-kubernetes-preserve-unknown-fields to be true")
	}

	// Description should still be set
	if schema.Description != "The principals this statement applies to" {
		t.Errorf("Expected description to be set, got '%s'", schema.Description)
	}
}

func TestGenerateXRDWithAnyTypeFields(t *testing.T) {
	// Test full XRD generation with 'any' type fields
	schema := &parser.Schema{
		Name: "TestSchema",
		Fields: []parser.Field{
			{
				Name:                  "principal",
				Type:                  "any",
				Required:              false,
				Description:           "Principal field",
				PreserveUnknownFields: true,
			},
			{
				Name:                  "action",
				Type:                  "any",
				Required:              false,
				Description:           "Action field",
				PreserveUnknownFields: true,
			},
			{
				Name:     "name",
				Type:     "str",
				Required: true,
			},
		},
	}

	xrdYAML, err := GenerateXRD(schema, "example.org", "v1alpha1")
	if err != nil {
		t.Fatalf("GenerateXRD failed: %v", err)
	}

	// Parse the YAML
	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Generated XRD is not valid YAML: %v", err)
	}

	// Navigate to parameters properties
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	versionSchema := version["schema"].(map[string]interface{})
	openAPISchema := versionSchema["openAPIV3Schema"].(map[string]interface{})
	properties := openAPISchema["properties"].(map[string]interface{})
	specProp := properties["spec"].(map[string]interface{})
	specProps := specProp["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})

	// Check principal field
	principal := paramProps["principal"].(map[string]interface{})
	if _, hasType := principal["type"]; hasType {
		t.Error("'any' type field should not have 'type' property")
	}
	if preserveUnknown := principal["x-kubernetes-preserve-unknown-fields"]; preserveUnknown != true {
		t.Error("Expected x-kubernetes-preserve-unknown-fields to be true for 'any' type")
	}
	if principal["description"] != "Principal field" {
		t.Errorf("Expected description 'Principal field', got '%v'", principal["description"])
	}

	// Check action field
	action := paramProps["action"].(map[string]interface{})
	if _, hasType := action["type"]; hasType {
		t.Error("'any' type field should not have 'type' property")
	}
	if preserveUnknown := action["x-kubernetes-preserve-unknown-fields"]; preserveUnknown != true {
		t.Error("Expected x-kubernetes-preserve-unknown-fields to be true for 'any' type")
	}

	// Check name field has type
	name := paramProps["name"].(map[string]interface{})
	if name["type"] != "string" {
		t.Errorf("Expected type 'string' for name field, got '%v'", name["type"])
	}
}
