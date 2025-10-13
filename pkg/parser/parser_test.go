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
