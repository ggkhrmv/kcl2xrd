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

func TestConvertFieldWithMapTypes(t *testing.T) {
	tests := []struct {
		name              string
		field             parser.Field
		expectedType      string
		expectedValueType string
	}{
		{
			name:              "string to string map",
			field:             parser.Field{Name: "labels", Type: "{str:str}"},
			expectedType:      "object",
			expectedValueType: "string",
		},
		{
			name:              "string to int map",
			field:             parser.Field{Name: "counts", Type: "{str:int}"},
			expectedType:      "object",
			expectedValueType: "integer",
		},
		{
			name:              "string to bool map",
			field:             parser.Field{Name: "flags", Type: "{str:bool}"},
			expectedType:      "object",
			expectedValueType: "boolean",
		},
		{
			name:              "string to float map",
			field:             parser.Field{Name: "metrics", Type: "{str:float}"},
			expectedType:      "object",
			expectedValueType: "number",
		},
		{
			name:              "any to any map",
			field:             parser.Field{Name: "config", Type: "{any:any}"},
			expectedType:      "object",
			expectedValueType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := convertFieldToPropertySchema(tt.field)

			if schema.Type != tt.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedType, schema.Type)
			}

			if schema.AdditionalProperties == nil {
				t.Error("Expected additionalProperties to be set for map type")
				return
			}

			if tt.expectedValueType == "" {
				// For {any:any}, additionalProperties should be an empty schema (allowing any type)
				if propSchema, ok := schema.AdditionalProperties.(*PropertySchema); ok && propSchema.Type != "" {
					t.Errorf("Expected empty additionalProperties type for {any:any}, got '%s'", propSchema.Type)
				}
			} else {
				if propSchema, ok := schema.AdditionalProperties.(*PropertySchema); ok {
					if propSchema.Type != tt.expectedValueType {
						t.Errorf("Expected additionalProperties type '%s', got '%s'", tt.expectedValueType, propSchema.Type)
					}
				} else {
					t.Errorf("Expected additionalProperties to be a PropertySchema")
				}
			}
		})
	}
}

func TestGenerateXRDWithMapTypes(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestMapSchema",
		Fields: []parser.Field{
			{
				Name:        "labels",
				Type:        "{str:str}",
				Required:    true,
				Description: "String to string map",
			},
			{
				Name:        "counts",
				Type:        "{str:int}",
				Required:    false,
				Description: "String to int map",
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

	// Check labels field
	labels := paramProps["labels"].(map[string]interface{})
	if labels["type"] != "object" {
		t.Errorf("Expected type 'object' for labels, got '%v'", labels["type"])
	}
	if labels["additionalProperties"] == nil {
		t.Error("Expected additionalProperties to be set for labels")
	} else {
		additionalProps := labels["additionalProperties"].(map[string]interface{})
		if additionalProps["type"] != "string" {
			t.Errorf("Expected additionalProperties type 'string' for labels, got '%v'", additionalProps["type"])
		}
	}

	// Check counts field
	counts := paramProps["counts"].(map[string]interface{})
	if counts["type"] != "object" {
		t.Errorf("Expected type 'object' for counts, got '%v'", counts["type"])
	}
	if counts["additionalProperties"] == nil {
		t.Error("Expected additionalProperties to be set for counts")
	} else {
		additionalProps := counts["additionalProperties"].(map[string]interface{})
		if additionalProps["type"] != "integer" {
			t.Errorf("Expected additionalProperties type 'integer' for counts, got '%v'", additionalProps["type"])
		}
	}
}

func TestGenerateXRDWithMinItems(t *testing.T) {
	// Test that @minItems annotation is properly applied to array fields
	minItems1 := 1
	minItems2 := 2
	schema := &parser.Schema{
		Name: "TestMinItems",
		Fields: []parser.Field{
			{
				Name:     "tags",
				Type:     "[str]",
				Required: true,
				MinItems: &minItems1,
			},
			{
				Name:     "items",
				Type:     "[str]",
				Required: false,
				MinItems: &minItems2,
				ListType: "set",
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

	// Check tags field
	tags := paramProps["tags"].(map[string]interface{})
	if tags["type"] != "array" {
		t.Errorf("Expected type 'array' for tags, got '%v'", tags["type"])
	}
	minItemsValue := tags["minItems"]
	if minItemsValue == nil {
		t.Error("Expected minItems to be set for tags field")
	} else if minItemsValue != 1 {
		t.Errorf("Expected minItems 1 for tags field, got %v", minItemsValue)
	}

	// Check items field
	items := paramProps["items"].(map[string]interface{})
	if items["type"] != "array" {
		t.Errorf("Expected type 'array' for items, got '%v'", items["type"])
	}
	minItemsValue = items["minItems"]
	if minItemsValue == nil {
		t.Error("Expected minItems to be set for items field")
	} else if minItemsValue != 2 {
		t.Errorf("Expected minItems 2 for items field, got %v", minItemsValue)
	}
	// Check that listType is also set
	listType := items["x-kubernetes-list-type"]
	if listType != "set" {
		t.Errorf("Expected x-kubernetes-list-type 'set', got '%v'", listType)
	}
}

func TestGenerateXRDWithOneOf(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestResource",
		Fields: []parser.Field{
			{
				Name:     "groupName",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "groupRef",
				Type:     "str",
				Required: false,
			},
			{
				Name: "config",
				Type: "{str:str}",
				OneOf: [][]string{
					{"groupName"},
					{"groupRef"},
				},
			},
		},
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, XRDOptions{
		Group:         "example.org",
		Version:       "v1",
		Served:        true,
		Referenceable: true,
	})
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Failed to unmarshal XRD: %v", err)
	}

	// Navigate to parameters schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})
	specProp := properties["spec"].(map[string]interface{})
	specProps := specProp["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})
	config := paramProps["config"].(map[string]interface{})

	// Check oneOf is present
	if config["oneOf"] == nil {
		t.Fatal("Expected oneOf to be set")
	}

	oneOf := config["oneOf"].([]interface{})
	if len(oneOf) != 2 {
		t.Errorf("Expected 2 oneOf schemas, got %d", len(oneOf))
	}

	// Check first oneOf entry
	oneOf0 := oneOf[0].(map[string]interface{})
	required0 := oneOf0["required"].([]interface{})
	if len(required0) != 1 || required0[0] != "groupName" {
		t.Errorf("Expected first oneOf to require 'groupName', got %v", required0)
	}

	// Check second oneOf entry
	oneOf1 := oneOf[1].(map[string]interface{})
	required1 := oneOf1["required"].([]interface{})
	if len(required1) != 1 || required1[0] != "groupRef" {
		t.Errorf("Expected second oneOf to require 'groupRef', got %v", required1)
	}
}

func TestGenerateXRDWithAnyOf(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestResource",
		Fields: []parser.Field{
			{
				Name:     "userEmail",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "userObjectId",
				Type:     "str",
				Required: false,
			},
			{
				Name: "userConfig",
				Type: "{str:str}",
				AnyOf: [][]string{
					{"userEmail"},
					{"userObjectId"},
				},
			},
		},
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, XRDOptions{
		Group:         "example.org",
		Version:       "v1",
		Served:        true,
		Referenceable: true,
	})
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Failed to unmarshal XRD: %v", err)
	}

	// Navigate to parameters schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})
	specProp := properties["spec"].(map[string]interface{})
	specProps := specProp["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})
	userConfig := paramProps["userConfig"].(map[string]interface{})

	// Check anyOf is present
	if userConfig["anyOf"] == nil {
		t.Fatal("Expected anyOf to be set")
	}

	anyOf := userConfig["anyOf"].([]interface{})
	if len(anyOf) != 2 {
		t.Errorf("Expected 2 anyOf schemas, got %d", len(anyOf))
	}

	// Check first anyOf entry
	anyOf0 := anyOf[0].(map[string]interface{})
	required0 := anyOf0["required"].([]interface{})
	if len(required0) != 1 || required0[0] != "userEmail" {
		t.Errorf("Expected first anyOf to require 'userEmail', got %v", required0)
	}

	// Check second anyOf entry
	anyOf1 := anyOf[1].(map[string]interface{})
	required1 := anyOf1["required"].([]interface{})
	if len(required1) != 1 || required1[0] != "userObjectId" {
		t.Errorf("Expected second anyOf to require 'userObjectId', got %v", required1)
	}
}

func TestGenerateXRDWithCombinedOneOfAndAnyOf(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestResource",
		Fields: []parser.Field{
			{
				Name:     "groupName",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "groupRef",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "userEmail",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "userObjectId",
				Type:     "str",
				Required: false,
			},
			{
				Name: "config",
				Type: "{str:str}",
				OneOf: [][]string{
					{"groupName"},
					{"groupRef"},
				},
				AnyOf: [][]string{
					{"userEmail"},
					{"userObjectId"},
				},
			},
		},
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, XRDOptions{
		Group:         "example.org",
		Version:       "v1",
		Served:        true,
		Referenceable: true,
	})
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Failed to unmarshal XRD: %v", err)
	}

	// Navigate to parameters schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})
	specProp := properties["spec"].(map[string]interface{})
	specProps := specProp["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})
	config := paramProps["config"].(map[string]interface{})

	// Check both oneOf and anyOf are present
	if config["oneOf"] == nil {
		t.Fatal("Expected oneOf to be set")
	}
	if config["anyOf"] == nil {
		t.Fatal("Expected anyOf to be set")
	}

	oneOf := config["oneOf"].([]interface{})
	if len(oneOf) != 2 {
		t.Errorf("Expected 2 oneOf schemas, got %d", len(oneOf))
	}

	anyOf := config["anyOf"].([]interface{})
	if len(anyOf) != 2 {
		t.Errorf("Expected 2 anyOf schemas, got %d", len(anyOf))
	}
}

func TestGenerateXRDWithSchemaLevelOneOf(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestResource",
		OneOf: [][]string{
			{"groupName"},
			{"groupRef"},
		},
		Fields: []parser.Field{
			{
				Name:     "groupName",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "groupRef",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "name",
				Type:     "str",
				Required: true,
			},
		},
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, XRDOptions{
		Group:         "example.org",
		Version:       "v1",
		Served:        true,
		Referenceable: true,
	})
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Failed to unmarshal XRD: %v", err)
	}

	// Navigate to parameters schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})
	specProp := properties["spec"].(map[string]interface{})
	specProps := specProp["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})

	// Check oneOf is present at parameters level
	if parameters["oneOf"] == nil {
		t.Fatal("Expected oneOf to be set at parameters level")
	}

	oneOf := parameters["oneOf"].([]interface{})
	if len(oneOf) != 2 {
		t.Errorf("Expected 2 oneOf schemas, got %d", len(oneOf))
	}

	// Check first oneOf entry
	oneOf0 := oneOf[0].(map[string]interface{})
	required0 := oneOf0["required"].([]interface{})
	if len(required0) != 1 || required0[0] != "groupName" {
		t.Errorf("Expected first oneOf to require 'groupName', got %v", required0)
	}

	// Check second oneOf entry
	oneOf1 := oneOf[1].(map[string]interface{})
	required1 := oneOf1["required"].([]interface{})
	if len(required1) != 1 || required1[0] != "groupRef" {
		t.Errorf("Expected second oneOf to require 'groupRef', got %v", required1)
	}
}

func TestGenerateXRDWithSchemaLevelAnyOf(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestResource",
		AnyOf: [][]string{
			{"userEmail"},
			{"userObjectId"},
		},
		Fields: []parser.Field{
			{
				Name:     "userEmail",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "userObjectId",
				Type:     "str",
				Required: false,
			},
			{
				Name:     "name",
				Type:     "str",
				Required: true,
			},
		},
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, XRDOptions{
		Group:         "example.org",
		Version:       "v1",
		Served:        true,
		Referenceable: true,
	})
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Failed to unmarshal XRD: %v", err)
	}

	// Navigate to parameters schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})
	specProp := properties["spec"].(map[string]interface{})
	specProps := specProp["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})

	// Check anyOf is present at parameters level
	if parameters["anyOf"] == nil {
		t.Fatal("Expected anyOf to be set at parameters level")
	}

	anyOf := parameters["anyOf"].([]interface{})
	if len(anyOf) != 2 {
		t.Errorf("Expected 2 anyOf schemas, got %d", len(anyOf))
	}

	// Check first anyOf entry
	anyOf0 := anyOf[0].(map[string]interface{})
	required0 := anyOf0["required"].([]interface{})
	if len(required0) != 1 || required0[0] != "userEmail" {
		t.Errorf("Expected first anyOf to require 'userEmail', got %v", required0)
	}

	// Check second anyOf entry
	anyOf1 := anyOf[1].(map[string]interface{})
	required1 := anyOf1["required"].([]interface{})
	if len(required1) != 1 || required1[0] != "userObjectId" {
		t.Errorf("Expected second anyOf to require 'userObjectId', got %v", required1)
	}
}

func TestGenerateXRDWithItemsFormat(t *testing.T) {
	schema := &parser.Schema{
		Name: "TestResource",
		Fields: []parser.Field{
			{
				Name:        "emails",
				Type:        "[str]",
				Required:    true,
				ItemsFormat: "email",
			},
			{
				Name:        "ids",
				Type:        "[str]",
				Required:    false,
				ItemsFormat: "uuid",
			},
		},
	}

	xrdYAML, err := GenerateXRDWithOptions(schema, XRDOptions{
		Group:         "example.org",
		Version:       "v1",
		Served:        true,
		Referenceable: true,
	})
	if err != nil {
		t.Fatalf("GenerateXRDWithOptions failed: %v", err)
	}

	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Failed to unmarshal XRD: %v", err)
	}

	// Navigate to parameters schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})
	specProp := properties["spec"].(map[string]interface{})
	specProps := specProp["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})

	// Check emails field
	emails := paramProps["emails"].(map[string]interface{})
	if emails["type"] != "array" {
		t.Errorf("Expected type 'array' for emails, got '%v'", emails["type"])
	}
	emailsItems := emails["items"].(map[string]interface{})
	if emailsItems["type"] != "string" {
		t.Errorf("Expected items type 'string' for emails, got '%v'", emailsItems["type"])
	}
	if emailsItems["format"] != "email" {
		t.Errorf("Expected items format 'email' for emails, got '%v'", emailsItems["format"])
	}

	// Check ids field
	ids := paramProps["ids"].(map[string]interface{})
	if ids["type"] != "array" {
		t.Errorf("Expected type 'array' for ids, got '%v'", ids["type"])
	}
	idsItems := ids["items"].(map[string]interface{})
	if idsItems["type"] != "string" {
		t.Errorf("Expected items type 'string' for ids, got '%v'", idsItems["type"])
	}
	if idsItems["format"] != "uuid" {
		t.Errorf("Expected items format 'uuid' for ids, got '%v'", idsItems["format"])
	}
}

func TestGenerateXRDWithStatusFields(t *testing.T) {
	schema := &parser.Schema{
		Name:        "MyResource",
		Description: "A resource with status fields",
		Fields: []parser.Field{
			{
				Name:     "name",
				Type:     "str",
				Required: true,
			},
			{
				Name:     "replicas",
				Type:     "int",
				Required: false,
				Default:  "3",
			},
			{
				Name:     "ready",
				Type:     "bool",
				Required: true,
				IsStatus: true,
			},
			{
				Name:     "phase",
				Type:     "str",
				Required: false,
				IsStatus: true,
			},
			{
				Name:                  "conditions",
				Type:                  "{any:any}",
				Required:              false,
				IsStatus:              true,
				PreserveUnknownFields: true,
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

	// Navigate to the schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})

	// Check that spec section exists with spec fields
	specSection := properties["spec"].(map[string]interface{})
	specProps := specSection["properties"].(map[string]interface{})
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})

	// Verify spec fields
	if _, ok := paramProps["name"]; !ok {
		t.Error("Expected 'name' field in spec.parameters")
	}
	if _, ok := paramProps["replicas"]; !ok {
		t.Error("Expected 'replicas' field in spec.parameters")
	}

	// Verify status fields are NOT in spec
	if _, ok := paramProps["ready"]; ok {
		t.Error("Status field 'ready' should not be in spec.parameters")
	}
	if _, ok := paramProps["phase"]; ok {
		t.Error("Status field 'phase' should not be in spec.parameters")
	}
	if _, ok := paramProps["conditions"]; ok {
		t.Error("Status field 'conditions' should not be in spec.parameters")
	}

	// Check that status section exists
	statusSection, ok := properties["status"]
	if !ok {
		t.Fatal("Expected 'status' section in openAPIV3Schema properties")
	}

	statusProps := statusSection.(map[string]interface{})["properties"].(map[string]interface{})

	// Verify status fields
	if _, ok := statusProps["ready"]; !ok {
		t.Error("Expected 'ready' field in status")
	}
	if _, ok := statusProps["phase"]; !ok {
		t.Error("Expected 'phase' field in status")
	}
	if _, ok := statusProps["conditions"]; !ok {
		t.Error("Expected 'conditions' field in status")
	}

	// Verify that preserveUnknownFields works on status fields
	conditions := statusProps["conditions"].(map[string]interface{})
	preserveUnknown := conditions["x-kubernetes-preserve-unknown-fields"]
	if preserveUnknown != true {
		t.Errorf("Expected x-kubernetes-preserve-unknown-fields true for conditions field, got %v", preserveUnknown)
	}

	// Verify spec fields are NOT in status
	if _, ok := statusProps["name"]; ok {
		t.Error("Spec field 'name' should not be in status")
	}
	if _, ok := statusProps["replicas"]; ok {
		t.Error("Spec field 'replicas' should not be in status")
	}

	// Check required fields
	statusRequired := statusSection.(map[string]interface{})["required"].([]interface{})
	hasReadyRequired := false
	for _, req := range statusRequired {
		if req.(string) == "ready" {
			hasReadyRequired = true
			break
		}
	}
	if !hasReadyRequired {
		t.Error("Expected 'ready' to be in status required fields")
	}
}

func TestGenerateXRDWithMaxItems(t *testing.T) {
	// Test that @maxItems annotation is properly applied to array fields
	maxItems1 := 5
	maxItems2 := 10
	minItems := 1
	schema := &parser.Schema{
		Name: "TestMaxItems",
		Fields: []parser.Field{
			{
				Name:     "tags",
				Type:     "[str]",
				Required: true,
				MaxItems: &maxItems1,
			},
			{
				Name:     "items",
				Type:     "[str]",
				Required: false,
				MinItems: &minItems,
				MaxItems: &maxItems2,
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

	// Check tags field
	tags := paramProps["tags"].(map[string]interface{})
	if tags["type"] != "array" {
		t.Errorf("Expected type 'array' for tags, got '%v'", tags["type"])
	}
	maxItemsValue := tags["maxItems"]
	if maxItemsValue == nil {
		t.Error("Expected maxItems to be set for tags field")
	} else if maxItemsValue != 5 {
		t.Errorf("Expected maxItems 5 for tags field, got %v", maxItemsValue)
	}

	// Check items field
	items := paramProps["items"].(map[string]interface{})
	if items["type"] != "array" {
		t.Errorf("Expected type 'array' for items, got '%v'", items["type"])
	}
	minItemsValue := items["minItems"]
	if minItemsValue == nil {
		t.Error("Expected minItems to be set for items field")
	} else if minItemsValue != 1 {
		t.Errorf("Expected minItems 1 for items field, got %v", minItemsValue)
	}
	maxItemsValue = items["maxItems"]
	if maxItemsValue == nil {
		t.Error("Expected maxItems to be set for items field")
	} else if maxItemsValue != 10 {
		t.Errorf("Expected maxItems 10 for items field, got %v", maxItemsValue)
	}
}

func TestGenerateXRDWithFormat(t *testing.T) {
	// Test that @format annotation is properly applied to string fields
	schema := &parser.Schema{
		Name: "TestFormat",
		Fields: []parser.Field{
			{
				Name:     "createdAt",
				Type:     "str",
				Required: true,
				Format:   "date-time",
			},
			{
				Name:   "email",
				Type:   "str",
				Format: "email",
				Pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			},
			{
				Name:   "id",
				Type:   "str",
				Format: "uuid",
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

	// Check createdAt field
	createdAt := paramProps["createdAt"].(map[string]interface{})
	if createdAt["type"] != "string" {
		t.Errorf("Expected type 'string' for createdAt, got '%v'", createdAt["type"])
	}
	if createdAt["format"] != "date-time" {
		t.Errorf("Expected format 'date-time' for createdAt, got '%v'", createdAt["format"])
	}

	// Check email field
	email := paramProps["email"].(map[string]interface{})
	if email["type"] != "string" {
		t.Errorf("Expected type 'string' for email, got '%v'", email["type"])
	}
	if email["format"] != "email" {
		t.Errorf("Expected format 'email' for email, got '%v'", email["format"])
	}
	if email["pattern"] == nil {
		t.Error("Expected pattern to be set for email field")
	}

	// Check id field
	id := paramProps["id"].(map[string]interface{})
	if id["type"] != "string" {
		t.Errorf("Expected type 'string' for id, got '%v'", id["type"])
	}
	if id["format"] != "uuid" {
		t.Errorf("Expected format 'uuid' for id, got '%v'", id["format"])
	}
}

func TestGenerateXRDWithSpecLevelFields(t *testing.T) {
	schema := &parser.Schema{
		Name:        "MyResource",
		Description: "A resource with spec-level fields",
		Fields: []parser.Field{
			{
				Name:     "name",
				Type:     "str",
				Required: true,
			},
			{
				Name:     "replicas",
				Type:     "int",
				Required: false,
				Default:  "3",
			},
			{
				Name:     "compositionSelector",
				Type:     "{str:str}",
				Required: false,
				IsSpec:   true,
			},
			{
				Name:     "customSpecField",
				Type:     "str",
				Required: true,
				IsSpec:   true,
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

	// Navigate to the schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})

	// Check that spec section exists
	specSection := properties["spec"].(map[string]interface{})
	specProps := specSection["properties"].(map[string]interface{})

	// Verify parameters section exists with regular fields
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})

	// Verify regular spec.parameters fields
	if _, ok := paramProps["name"]; !ok {
		t.Error("Expected 'name' field in spec.parameters")
	}
	if _, ok := paramProps["replicas"]; !ok {
		t.Error("Expected 'replicas' field in spec.parameters")
	}

	// Verify spec-level fields are NOT in parameters
	if _, ok := paramProps["compositionSelector"]; ok {
		t.Error("Spec-level field 'compositionSelector' should not be in spec.parameters")
	}
	if _, ok := paramProps["customSpecField"]; ok {
		t.Error("Spec-level field 'customSpecField' should not be in spec.parameters")
	}

	// Verify spec-level fields exist directly under spec
	if _, ok := specProps["compositionSelector"]; !ok {
		t.Error("Expected 'compositionSelector' field directly under spec")
	}
	if _, ok := specProps["customSpecField"]; !ok {
		t.Error("Expected 'customSpecField' field directly under spec")
	}

	// Verify spec required includes the required spec-level field
	specRequired := specSection["required"].([]interface{})
	foundCustomSpecField := false
	foundParameters := false
	for _, req := range specRequired {
		if req == "customSpecField" {
			foundCustomSpecField = true
		}
		if req == "parameters" {
			foundParameters = true
		}
	}
	if !foundCustomSpecField {
		t.Error("Expected 'customSpecField' to be in spec.required")
	}
	if !foundParameters {
		t.Error("Expected 'parameters' to be in spec.required")
	}

	// Verify the type of compositionSelector
	compositionSelector := specProps["compositionSelector"].(map[string]interface{})
	if compositionSelector["type"] != "object" {
		t.Errorf("Expected type 'object' for compositionSelector, got '%v'", compositionSelector["type"])
	}
}

func TestGenerateXRDWithSpecPathSchemas(t *testing.T) {
	// Main schema
	mainSchema := &parser.Schema{
		Name:        "MyApp",
		Description: "Main application schema",
		Fields: []parser.Field{
			{
				Name:     "name",
				Type:     "str",
				Required: true,
			},
			{
				Name:     "replicas",
				Type:     "int",
				Required: false,
				Default:  "3",
			},
		},
	}

	// Schema for spec.customParameters
	customParamsSchema := &parser.Schema{
		Name:     "CustomParams",
		SpecPath: "customParameters",
		Fields: []parser.Field{
			{
				Name:     "customField1",
				Type:     "str",
				Required: true,
			},
			{
				Name:     "customField2",
				Type:     "int",
				Required: false,
			},
		},
	}

	// Schema for spec.writeConnectionSecretToRef
	connectionSecretSchema := &parser.Schema{
		Name:     "ConnectionSecret",
		SpecPath: "writeConnectionSecretToRef",
		Fields: []parser.Field{
			{
				Name:     "name",
				Type:     "str",
				Required: true,
			},
			{
				Name:     "namespace",
				Type:     "str",
				Required: false,
			},
		},
	}

	// Status schema
	statusSchema := &parser.Schema{
		Name:     "AppStatus",
		IsStatus: true,
		Fields: []parser.Field{
			{
				Name:     "ready",
				Type:     "bool",
				Required: true,
			},
		},
	}

	schemas := map[string]*parser.Schema{
		"MyApp":            mainSchema,
		"CustomParams":     customParamsSchema,
		"ConnectionSecret": connectionSecretSchema,
		"AppStatus":        statusSchema,
	}

	xrdYAML, err := GenerateXRDWithSchemasAndOptions(mainSchema, schemas, XRDOptions{
		Group:   "example.org",
		Version: "v1alpha1",
	})
	if err != nil {
		t.Fatalf("GenerateXRDWithSchemasAndOptions failed: %v", err)
	}

	// Check that it's valid YAML
	var xrd map[string]interface{}
	if err := yaml.Unmarshal([]byte(xrdYAML), &xrd); err != nil {
		t.Fatalf("Generated XRD is not valid YAML: %v", err)
	}

	// Navigate to the schema
	spec := xrd["spec"].(map[string]interface{})
	versions := spec["versions"].([]interface{})
	version := versions[0].(map[string]interface{})
	schema_obj := version["schema"].(map[string]interface{})
	openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
	properties := openAPIV3Schema["properties"].(map[string]interface{})

	// Check that spec section exists
	specSection := properties["spec"].(map[string]interface{})
	specProps := specSection["properties"].(map[string]interface{})

	// Verify parameters section exists with regular fields
	parameters := specProps["parameters"].(map[string]interface{})
	paramProps := parameters["properties"].(map[string]interface{})

	// Verify regular spec.parameters fields
	if _, ok := paramProps["name"]; !ok {
		t.Error("Expected 'name' field in spec.parameters")
	}
	if _, ok := paramProps["replicas"]; !ok {
		t.Error("Expected 'replicas' field in spec.parameters")
	}

	// Verify spec path schemas are in spec (not in parameters)
	if _, ok := specProps["customParameters"]; !ok {
		t.Error("Expected 'customParameters' in spec properties")
	}
	if _, ok := specProps["writeConnectionSecretToRef"]; !ok {
		t.Error("Expected 'writeConnectionSecretToRef' in spec properties")
	}

	// Check customParameters fields
	customParams := specProps["customParameters"].(map[string]interface{})
	customParamsProps := customParams["properties"].(map[string]interface{})
	if _, ok := customParamsProps["customField1"]; !ok {
		t.Error("Expected 'customField1' in customParameters")
	}
	if _, ok := customParamsProps["customField2"]; !ok {
		t.Error("Expected 'customField2' in customParameters")
	}

	// Check customParameters required
	customParamsRequired := customParams["required"].([]interface{})
	if len(customParamsRequired) != 1 || customParamsRequired[0] != "customField1" {
		t.Errorf("Expected customParameters.required to contain 'customField1', got %v", customParamsRequired)
	}

	// Check writeConnectionSecretToRef fields
	connectionSecret := specProps["writeConnectionSecretToRef"].(map[string]interface{})
	connectionSecretProps := connectionSecret["properties"].(map[string]interface{})
	if _, ok := connectionSecretProps["name"]; !ok {
		t.Error("Expected 'name' in writeConnectionSecretToRef")
	}
	if _, ok := connectionSecretProps["namespace"]; !ok {
		t.Error("Expected 'namespace' in writeConnectionSecretToRef")
	}

	// Verify status section exists
	statusSection := properties["status"].(map[string]interface{})
	statusProps := statusSection["properties"].(map[string]interface{})
	if _, ok := statusProps["ready"]; !ok {
		t.Error("Expected 'ready' field in status")
	}
}



func TestGenerateXRDWithItemsPreserveUnknownFields(t *testing.T) {
schema := &parser.Schema{
Name: "XMyApp",
Fields: []parser.Field{
{
Name:     "name",
Type:     "str",
Required: true,
},
{
Name:                       "filter",
Type:                       "[{any:any}]",
Required:                   true,
PreserveUnknownFields:      true,
},
{
Name:                       "configs",
Type:                       "[{str:str}]",
Required:                   true,
ItemsPreserveUnknownFields: true,
},
{
Name:                  "metadata",
Type:                  "{any:any}",
Required:              false,
PreserveUnknownFields: true,
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

// Navigate to the schema
spec := xrd["spec"].(map[string]interface{})
versions := spec["versions"].([]interface{})
version := versions[0].(map[string]interface{})
schema_obj := version["schema"].(map[string]interface{})
openAPIV3Schema := schema_obj["openAPIV3Schema"].(map[string]interface{})
properties := openAPIV3Schema["properties"].(map[string]interface{})
specProp := properties["spec"].(map[string]interface{})
specProps := specProp["properties"].(map[string]interface{})
parameters := specProps["parameters"].(map[string]interface{})
paramProps := parameters["properties"].(map[string]interface{})

// Check filter field - should have x-kubernetes-preserve-unknown-fields only on items
filter := paramProps["filter"].(map[string]interface{})
if filter["type"] != "array" {
t.Errorf("Expected type 'array' for filter, got '%v'", filter["type"])
}

// Filter should NOT have x-kubernetes-preserve-unknown-fields at array level
if _, hasPreserve := filter["x-kubernetes-preserve-unknown-fields"]; hasPreserve {
t.Error("Array 'filter' should not have x-kubernetes-preserve-unknown-fields")
}

// Filter items SHOULD have x-kubernetes-preserve-unknown-fields
filterItems := filter["items"].(map[string]interface{})
if filterItems["type"] != "object" {
t.Errorf("Expected items type 'object' for filter, got '%v'", filterItems["type"])
}
if preserveVal, ok := filterItems["x-kubernetes-preserve-unknown-fields"]; !ok || preserveVal != true {
t.Error("Filter items should have x-kubernetes-preserve-unknown-fields: true")
}

// Check configs field - should have x-kubernetes-preserve-unknown-fields only on items
configs := paramProps["configs"].(map[string]interface{})
if configs["type"] != "array" {
t.Errorf("Expected type 'array' for configs, got '%v'", configs["type"])
}

// Configs should NOT have x-kubernetes-preserve-unknown-fields at array level
if _, hasPreserve := configs["x-kubernetes-preserve-unknown-fields"]; hasPreserve {
t.Error("Array 'configs' should not have x-kubernetes-preserve-unknown-fields")
}

// Configs items SHOULD have x-kubernetes-preserve-unknown-fields
configsItems := configs["items"].(map[string]interface{})
if configsItems["type"] != "object" {
t.Errorf("Expected items type 'object' for configs, got '%v'", configsItems["type"])
}
if preserveVal, ok := configsItems["x-kubernetes-preserve-unknown-fields"]; !ok || preserveVal != true {
t.Error("Configs items should have x-kubernetes-preserve-unknown-fields: true")
}

// Check metadata field - should have x-kubernetes-preserve-unknown-fields on the object itself
metadata := paramProps["metadata"].(map[string]interface{})
if metadata["type"] != "object" {
t.Errorf("Expected type 'object' for metadata, got '%v'", metadata["type"])
}
if preserveVal, ok := metadata["x-kubernetes-preserve-unknown-fields"]; !ok || preserveVal != true {
t.Error("Metadata object should have x-kubernetes-preserve-unknown-fields: true")
}
}
