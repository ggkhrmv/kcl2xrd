package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseKCLFile(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")

	content := `schema TestSchema:
    r"""
    Test schema description
    """
    
    requiredField: int
    
    optionalField?: str
    
    defaultField?: str = "default"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	schema, err := ParseKCLFile(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFile failed: %v", err)
	}

	// Check schema name
	if schema.Name != "TestSchema" {
		t.Errorf("Expected schema name 'TestSchema', got '%s'", schema.Name)
	}

	// Check number of fields
	if len(schema.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(schema.Fields))
	}

	// Check required field
	if len(schema.Fields) > 0 {
		field := schema.Fields[0]
		if field.Name != "requiredField" {
			t.Errorf("Expected first field name 'requiredField', got '%s'", field.Name)
		}
		if field.Type != "int" {
			t.Errorf("Expected first field type 'int', got '%s'", field.Type)
		}
		if !field.Required {
			t.Errorf("Expected first field to be required")
		}
	}

	// Check optional field
	if len(schema.Fields) > 1 {
		field := schema.Fields[1]
		if field.Name != "optionalField" {
			t.Errorf("Expected second field name 'optionalField', got '%s'", field.Name)
		}
		if field.Type != "str" {
			t.Errorf("Expected second field type 'str', got '%s'", field.Type)
		}
		if field.Required {
			t.Errorf("Expected second field to be optional")
		}
	}

	// Check default field
	if len(schema.Fields) > 2 {
		field := schema.Fields[2]
		if field.Name != "defaultField" {
			t.Errorf("Expected third field name 'defaultField', got '%s'", field.Name)
		}
		if field.Default == "" {
			t.Errorf("Expected third field to have a default value")
		}
	}
}

func TestParseKCLFileNotFound(t *testing.T) {
	_, err := ParseKCLFile("/nonexistent/file.k")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestParseKCLFileNoSchema(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")

	content := `# This file has no schema
import something
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseKCLFile(testFile)
	if err == nil {
		t.Error("Expected error for file with no schema, got nil")
	}
}

func TestParseKCLFileWithValidations(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")
	
	content := `schema TestSchema:
    # @pattern("^[a-z]+$")
    # @minLength(3)
    # @maxLength(10)
    name: str
    
    # @minimum(0)
    # @maximum(100)
    age?: int
    
    # @enum(["active", "inactive"])
    status?: str
    
    # @immutable
    id: str
    
    # @validate("self > 0", "Must be positive")
    count?: int
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	schema, err := ParseKCLFile(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFile failed: %v", err)
	}
	
	// Check name field validations
	nameField := schema.Fields[0]
	if nameField.Pattern != "^[a-z]+$" {
		t.Errorf("Expected pattern '^[a-z]+$', got '%s'", nameField.Pattern)
	}
	if nameField.MinLength == nil || *nameField.MinLength != 3 {
		t.Error("Expected minLength of 3")
	}
	if nameField.MaxLength == nil || *nameField.MaxLength != 10 {
		t.Error("Expected maxLength of 10")
	}
	
	// Check age field validations
	ageField := schema.Fields[1]
	if ageField.Minimum == nil || *ageField.Minimum != 0 {
		t.Error("Expected minimum of 0")
	}
	if ageField.Maximum == nil || *ageField.Maximum != 100 {
		t.Error("Expected maximum of 100")
	}
	
	// Check status field enum
	statusField := schema.Fields[2]
	if len(statusField.Enum) != 2 {
		t.Errorf("Expected 2 enum values, got %d", len(statusField.Enum))
	}
	
	// Check id field immutability
	idField := schema.Fields[3]
	if !idField.Immutable {
		t.Error("Expected id field to be immutable")
	}
	
	// Check count field CEL validation
	countField := schema.Fields[4]
	if len(countField.CELValidations) != 1 {
		t.Errorf("Expected 1 CEL validation, got %d", len(countField.CELValidations))
	}
	if countField.CELValidations[0].Rule != "self > 0" {
		t.Errorf("Expected CEL rule 'self > 0', got '%s'", countField.CELValidations[0].Rule)
	}
}

func TestParseKCLFileWithNestedSchemas(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")
	
	content := `schema NestedSchema:
    field1: str
    field2: int

schema MainSchema:
    nested: NestedSchema
    name: str
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result, err := ParseKCLFileWithSchemas(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
	}
	
	// Check that we have 2 schemas
	if len(result.Schemas) != 2 {
		t.Errorf("Expected 2 schemas, got %d", len(result.Schemas))
	}
	
	// Check that both schemas exist
	if result.Schemas["NestedSchema"] == nil {
		t.Error("Expected NestedSchema to be parsed")
	}
	if result.Schemas["MainSchema"] == nil {
		t.Error("Expected MainSchema to be parsed")
	}
	
	// Check that primary schema is MainSchema (last one)
	if result.Primary.Name != "MainSchema" {
		t.Errorf("Expected primary schema to be 'MainSchema', got '%s'", result.Primary.Name)
	}
	
	// Check nested schema fields
	nestedSchema := result.Schemas["NestedSchema"]
	if len(nestedSchema.Fields) != 2 {
		t.Errorf("Expected NestedSchema to have 2 fields, got %d", len(nestedSchema.Fields))
	}
	
	// Check main schema fields
	mainSchema := result.Schemas["MainSchema"]
	if len(mainSchema.Fields) != 2 {
		t.Errorf("Expected MainSchema to have 2 fields, got %d", len(mainSchema.Fields))
	}
	
	// Check that nested field has correct type
	if mainSchema.Fields[0].Type != "NestedSchema" {
		t.Errorf("Expected nested field type 'NestedSchema', got '%s'", mainSchema.Fields[0].Type)
	}
}

func TestParseKCLFileWithModuleLevelVariables(t *testing.T) {
	// Test that non-indented module-level variables after schemas don't get parsed as fields
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")
	
	content := `schema BucketPolicyStatement:
    # @enum(["Allow", "Deny"])
    effect?: str
    # @preserveUnknownFields
    principal?: any
    # @preserveUnknownFields
    action?: any

_xrSubgroup = "aws"

_composition: schemas.Composition{
    xrKind: "Test"
}
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result, err := ParseKCLFileWithSchemas(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
	}
	
	schema := result.Schemas["BucketPolicyStatement"]
	if schema == nil {
		t.Fatal("Expected BucketPolicyStatement schema to be parsed")
	}
	
	// Should have exactly 3 fields (effect, principal, action)
	if len(schema.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(schema.Fields))
		for i, field := range schema.Fields {
			t.Logf("Field %d: %s (%s)", i, field.Name, field.Type)
		}
	}
	
	// Verify we don't have _xrSubgroup or _composition fields
	for _, field := range schema.Fields {
		if field.Name == "_xrSubgroup" || field.Name == "_composition" || 
		   field.Name == "xrKind" || field.Name == "selector" {
			t.Errorf("Field '%s' should not be part of schema - module-level variables should not be parsed as fields", field.Name)
		}
	}
}

func TestParseKCLFileWithGroupExpression(t *testing.T) {
	// Test that __xrd_group with format expressions can be resolved
	tempDir := t.TempDir()
	
	t.Run("resolvable_format_expression", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test1.k")
		content := `_xrSubgroup = "aws"
_platformGroup = "example.org"

__xrd_kind = "Bucket"
__xrd_group = "{}.{}".format(_xrSubgroup, _platformGroup)
__xrd_version = "v1alpha1"

schema Bucket:
    name: str
`
		
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		result, err := ParseKCLFileWithSchemas(testFile)
		if err != nil {
			t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
		}
		
		// Group should be resolved from the format expression
		expected := "aws.example.org"
		if result.Metadata.Group != expected {
			t.Errorf("Expected Group '%s', got '%s'", expected, result.Metadata.Group)
		}
		
		// XRKind and XRVersion should also be parsed
		if result.Metadata.XRKind != "Bucket" {
			t.Errorf("Expected XRKind 'Bucket', got '%s'", result.Metadata.XRKind)
		}
		if result.Metadata.XRVersion != "v1alpha1" {
			t.Errorf("Expected XRVersion 'v1alpha1', got '%s'", result.Metadata.XRVersion)
		}
	})
	
	t.Run("unresolvable_format_expression", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test2.k")
		content := `__xrd_kind = "Bucket"
__xrd_group = "{}.{}".format(_xrSubgroup, settings.PLATFORM_API_GROUP)
__xrd_version = "v1alpha1"

schema Bucket:
    name: str
`
		
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		result, err := ParseKCLFileWithSchemas(testFile)
		if err != nil {
			t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
		}
		
		// Group should be empty since variables are not defined
		if result.Metadata.Group != "" {
			t.Errorf("Expected Group to be empty for unresolvable expression, got '%s'", result.Metadata.Group)
		}
		
		// But XRKind and XRVersion should be parsed
		if result.Metadata.XRKind != "Bucket" {
			t.Errorf("Expected XRKind 'Bucket', got '%s'", result.Metadata.XRKind)
		}
		if result.Metadata.XRVersion != "v1alpha1" {
			t.Errorf("Expected XRVersion 'v1alpha1', got '%s'", result.Metadata.XRVersion)
		}
	})
}

func TestParseKCLFileWithAnyType(t *testing.T) {
	// Test that fields with 'any' type are properly parsed
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")
	
	content := `schema TestSchema:
    # @preserveUnknownFields
    # Description for principal
    principal?: any
    
    # @preserveUnknownFields
    # Description for action
    action?: any
    
    name: str
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result, err := ParseKCLFileWithSchemas(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
	}
	
	schema := result.Schemas["TestSchema"]
	if schema == nil {
		t.Fatal("Expected TestSchema to be parsed")
	}
	
	// Check principal field
	if len(schema.Fields) < 1 {
		t.Fatal("Expected at least 1 field")
	}
	principalField := schema.Fields[0]
	if principalField.Name != "principal" {
		t.Errorf("Expected field name 'principal', got '%s'", principalField.Name)
	}
	if principalField.Type != "any" {
		t.Errorf("Expected type 'any', got '%s'", principalField.Type)
	}
	if !principalField.PreserveUnknownFields {
		t.Error("Expected PreserveUnknownFields to be true")
	}
	if principalField.Description == "" {
		t.Error("Expected description to be set")
	}
	
	// Check action field
	if len(schema.Fields) < 2 {
		t.Fatal("Expected at least 2 fields")
	}
	actionField := schema.Fields[1]
	if actionField.Name != "action" {
		t.Errorf("Expected field name 'action', got '%s'", actionField.Name)
	}
	if actionField.Type != "any" {
		t.Errorf("Expected type 'any', got '%s'", actionField.Type)
	}
	if !actionField.PreserveUnknownFields {
		t.Error("Expected PreserveUnknownFields to be true")
	}
}

func TestParseKCLFileWithMinItems(t *testing.T) {
	// Test that @minItems annotation is properly parsed
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")
	
	content := `schema TestSchema:
    # @minItems(1)
    tags: [str]
    
    # @minItems(2)
    # @listType("set")
    items?: [str]
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result, err := ParseKCLFileWithSchemas(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
	}
	
	schema := result.Schemas["TestSchema"]
	if schema == nil {
		t.Fatal("Expected TestSchema to be parsed")
	}
	
	// Check tags field
	if len(schema.Fields) < 1 {
		t.Fatal("Expected at least 1 field")
	}
	tagsField := schema.Fields[0]
	if tagsField.Name != "tags" {
		t.Errorf("Expected field name 'tags', got '%s'", tagsField.Name)
	}
	if tagsField.MinItems == nil || *tagsField.MinItems != 1 {
		t.Error("Expected minItems of 1 for tags field")
	}
	
	// Check items field
	if len(schema.Fields) < 2 {
		t.Fatal("Expected at least 2 fields")
	}
	itemsField := schema.Fields[1]
	if itemsField.Name != "items" {
		t.Errorf("Expected field name 'items', got '%s'", itemsField.Name)
	}
	if itemsField.MinItems == nil || *itemsField.MinItems != 2 {
		t.Error("Expected minItems of 2 for items field")
	}
	if itemsField.ListType != "set" {
		t.Errorf("Expected listType 'set', got '%s'", itemsField.ListType)
	}
}

func TestParseKCLFileWithStatusAnnotation(t *testing.T) {
	// Create a temporary test file with status annotations
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_status.k")

	content := `schema TestResource:
    # Regular spec field
    name: str
    
    # @status
    ready: bool
    
    # @status
    # @preserveUnknownFields
    conditions?: {any:any}
    
    # @status
    phase?: str
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	result, err := ParseKCLFileWithSchemas(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
	}

	schema := result.Primary
	if schema == nil {
		t.Fatal("Expected primary schema to be set")
	}

	// Check number of fields
	if len(schema.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(schema.Fields))
	}

	// Check name field (should NOT be status)
	nameField := schema.Fields[0]
	if nameField.Name != "name" {
		t.Errorf("Expected field name 'name', got '%s'", nameField.Name)
	}
	if nameField.IsStatus {
		t.Error("Expected 'name' field to NOT be a status field")
	}

	// Check ready field (should be status)
	readyField := schema.Fields[1]
	if readyField.Name != "ready" {
		t.Errorf("Expected field name 'ready', got '%s'", readyField.Name)
	}
	if !readyField.IsStatus {
		t.Error("Expected 'ready' field to be a status field")
	}

	// Check conditions field (should be status with preserveUnknownFields)
	conditionsField := schema.Fields[2]
	if conditionsField.Name != "conditions" {
		t.Errorf("Expected field name 'conditions', got '%s'", conditionsField.Name)
	}
	if !conditionsField.IsStatus {
		t.Error("Expected 'conditions' field to be a status field")
	}
	if !conditionsField.PreserveUnknownFields {
		t.Error("Expected 'conditions' field to have preserveUnknownFields set")
	}

	// Check phase field (should be status)
	phaseField := schema.Fields[3]
	if phaseField.Name != "phase" {
		t.Errorf("Expected field name 'phase', got '%s'", phaseField.Name)
	}
	if !phaseField.IsStatus {
		t.Error("Expected 'phase' field to be a status field")
	}
}

func TestParseKCLFileWithMaxItems(t *testing.T) {
	// Test that @maxItems annotation is properly parsed
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")
	
	content := `schema TestSchema:
    # @maxItems(5)
    tags: [str]
    
    # @minItems(1)
    # @maxItems(10)
    items?: [str]
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result, err := ParseKCLFileWithSchemas(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
	}
	
	schema := result.Schemas["TestSchema"]
	if schema == nil {
		t.Fatal("Expected TestSchema to be parsed")
	}
	
	// Check tags field
	if len(schema.Fields) < 1 {
		t.Fatal("Expected at least 1 field")
	}
	tagsField := schema.Fields[0]
	if tagsField.Name != "tags" {
		t.Errorf("Expected field name 'tags', got '%s'", tagsField.Name)
	}
	if tagsField.MaxItems == nil || *tagsField.MaxItems != 5 {
		t.Error("Expected maxItems of 5 for tags field")
	}
	
	// Check items field
	if len(schema.Fields) < 2 {
		t.Fatal("Expected at least 2 fields")
	}
	itemsField := schema.Fields[1]
	if itemsField.Name != "items" {
		t.Errorf("Expected field name 'items', got '%s'", itemsField.Name)
	}
	if itemsField.MinItems == nil || *itemsField.MinItems != 1 {
		t.Error("Expected minItems of 1 for items field")
	}
	if itemsField.MaxItems == nil || *itemsField.MaxItems != 10 {
		t.Error("Expected maxItems of 10 for items field")
	}
}

func TestParseKCLFileWithFormat(t *testing.T) {
	// Test that @format annotation is properly parsed
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")
	
	content := `schema TestSchema:
    # @format("date-time")
    createdAt: str
    
    # @format("email")
    # @pattern("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
    email: str
    
    # @format("uuid")
    id: str
`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result, err := ParseKCLFileWithSchemas(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFileWithSchemas failed: %v", err)
	}
	
	schema := result.Schemas["TestSchema"]
	if schema == nil {
		t.Fatal("Expected TestSchema to be parsed")
	}
	
	// Check createdAt field
	if len(schema.Fields) < 1 {
		t.Fatal("Expected at least 1 field")
	}
	createdAtField := schema.Fields[0]
	if createdAtField.Name != "createdAt" {
		t.Errorf("Expected field name 'createdAt', got '%s'", createdAtField.Name)
	}
	if createdAtField.Format != "date-time" {
		t.Errorf("Expected format 'date-time' for createdAt field, got '%s'", createdAtField.Format)
	}
	
	// Check email field
	if len(schema.Fields) < 2 {
		t.Fatal("Expected at least 2 fields")
	}
	emailField := schema.Fields[1]
	if emailField.Name != "email" {
		t.Errorf("Expected field name 'email', got '%s'", emailField.Name)
	}
	if emailField.Format != "email" {
		t.Errorf("Expected format 'email' for email field, got '%s'", emailField.Format)
	}
	if emailField.Pattern == "" {
		t.Error("Expected pattern to be set for email field")
	}
	
	// Check id field
	if len(schema.Fields) < 3 {
		t.Fatal("Expected at least 3 fields")
	}
	idField := schema.Fields[2]
	if idField.Name != "id" {
		t.Errorf("Expected field name 'id', got '%s'", idField.Name)
	}
	if idField.Format != "uuid" {
		t.Errorf("Expected format 'uuid' for id field, got '%s'", idField.Format)
	}
}

func TestParseKCLFileWithOneOf(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")

	content := `schema TestSchema:
    groupName?: str
    groupRef?: str
    
    # @oneOf([["groupName"], ["groupRef"]])
    config: {str:str}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	schema, err := ParseKCLFile(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFile failed: %v", err)
	}

	// Check oneOf field
	if len(schema.Fields) < 3 {
		t.Fatal("Expected at least 3 fields")
	}
	
	configField := schema.Fields[2]
	if configField.Name != "config" {
		t.Errorf("Expected field name 'config', got '%s'", configField.Name)
	}
	
	if len(configField.OneOf) != 2 {
		t.Fatalf("Expected 2 oneOf combinations, got %d", len(configField.OneOf))
	}
	
	// Check first oneOf combination
	if len(configField.OneOf[0]) != 1 || configField.OneOf[0][0] != "groupName" {
		t.Errorf("Expected first oneOf to be ['groupName'], got %v", configField.OneOf[0])
	}
	
	// Check second oneOf combination
	if len(configField.OneOf[1]) != 1 || configField.OneOf[1][0] != "groupRef" {
		t.Errorf("Expected second oneOf to be ['groupRef'], got %v", configField.OneOf[1])
	}
}

func TestParseKCLFileWithAnyOf(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")

	content := `schema TestSchema:
    userEmail?: str
    userObjectId?: str
    
    # @anyOf([["userEmail"], ["userObjectId"]])
    userConfig: {str:str}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	schema, err := ParseKCLFile(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFile failed: %v", err)
	}

	// Check anyOf field
	if len(schema.Fields) < 3 {
		t.Fatal("Expected at least 3 fields")
	}
	
	userConfigField := schema.Fields[2]
	if userConfigField.Name != "userConfig" {
		t.Errorf("Expected field name 'userConfig', got '%s'", userConfigField.Name)
	}
	
	if len(userConfigField.AnyOf) != 2 {
		t.Fatalf("Expected 2 anyOf combinations, got %d", len(userConfigField.AnyOf))
	}
	
	// Check first anyOf combination
	if len(userConfigField.AnyOf[0]) != 1 || userConfigField.AnyOf[0][0] != "userEmail" {
		t.Errorf("Expected first anyOf to be ['userEmail'], got %v", userConfigField.AnyOf[0])
	}
	
	// Check second anyOf combination
	if len(userConfigField.AnyOf[1]) != 1 || userConfigField.AnyOf[1][0] != "userObjectId" {
		t.Errorf("Expected second anyOf to be ['userObjectId'], got %v", userConfigField.AnyOf[1])
	}
}

func TestParseKCLFileWithCombinedOneOfAndAnyOf(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.k")

	content := `schema TestSchema:
    groupName?: str
    groupRef?: str
    userEmail?: str
    userObjectId?: str
    
    # @oneOf([["groupName"], ["groupRef"]])
    # @anyOf([["userEmail"], ["userObjectId"]])
    config: {str:str}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	schema, err := ParseKCLFile(testFile)
	if err != nil {
		t.Fatalf("ParseKCLFile failed: %v", err)
	}

	// Check field with both oneOf and anyOf
	if len(schema.Fields) < 5 {
		t.Fatal("Expected at least 5 fields")
	}
	
	configField := schema.Fields[4]
	if configField.Name != "config" {
		t.Errorf("Expected field name 'config', got '%s'", configField.Name)
	}
	
	// Check oneOf
	if len(configField.OneOf) != 2 {
		t.Fatalf("Expected 2 oneOf combinations, got %d", len(configField.OneOf))
	}
	
	// Check anyOf
	if len(configField.AnyOf) != 2 {
		t.Fatalf("Expected 2 anyOf combinations, got %d", len(configField.AnyOf))
	}
}


