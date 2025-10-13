package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ggkhrmv/kcl2xrd/pkg/parser"
	"gopkg.in/yaml.v3"
)

// XRD represents a Crossplane Composite Resource Definition
type XRD struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       XRDSpec     `yaml:"spec"`
}

// Metadata represents the metadata section of an XRD
type Metadata struct {
	Name string `yaml:"name"`
}

// XRDSpec represents the spec section of an XRD
type XRDSpec struct {
	Group      string      `yaml:"group"`
	Names      Names       `yaml:"names"`
	ClaimNames *ClaimNames `yaml:"claimNames,omitempty"`
	Versions   []Version   `yaml:"versions"`
}

// Names represents the names section of an XRD spec
type Names struct {
	Kind   string `yaml:"kind"`
	Plural string `yaml:"plural"`
}

// ClaimNames represents optional claim names in an XRD spec
type ClaimNames struct {
	Kind   string `yaml:"kind"`
	Plural string `yaml:"plural"`
}

// XRDOptions contains options for generating an XRD
type XRDOptions struct {
	Group       string
	Version     string
	WithClaims  bool
	ClaimKind   string
	ClaimPlural string
}

// Version represents a version in an XRD spec
type Version struct {
	Name          string        `yaml:"name"`
	Served        bool          `yaml:"served"`
	Referenceable bool          `yaml:"referenceable"`
	Schema        VersionSchema `yaml:"schema"`
}

// VersionSchema represents the schema section of a version
type VersionSchema struct {
	OpenAPIV3Schema OpenAPIV3Schema `yaml:"openAPIV3Schema"`
}

// OpenAPIV3Schema represents an OpenAPI v3 schema
type OpenAPIV3Schema struct {
	Type       string                     `yaml:"type"`
	Properties map[string]PropertySchema  `yaml:"properties,omitempty"`
	Required   []string                   `yaml:"required,omitempty"`
}

// PropertySchema represents a property in an OpenAPI schema
type PropertySchema struct {
	Type        string                    `yaml:"type,omitempty"`
	Description string                    `yaml:"description,omitempty"`
	Properties  map[string]PropertySchema `yaml:"properties,omitempty"`
	Required    []string                  `yaml:"required,omitempty"`
	Items       *PropertySchema           `yaml:"items,omitempty"`
	Format      string                    `yaml:"format,omitempty"`
	Default     interface{}               `yaml:"default,omitempty"`
}

// GenerateXRD generates a Crossplane XRD from a parsed KCL schema
// Deprecated: Use GenerateXRDWithOptions for more control
func GenerateXRD(schema *parser.Schema, group, version string) (string, error) {
	return GenerateXRDWithOptions(schema, XRDOptions{
		Group:   group,
		Version: version,
	})
}

// GenerateXRDWithOptions generates a Crossplane XRD from a parsed KCL schema with options
func GenerateXRDWithOptions(schema *parser.Schema, opts XRDOptions) (string, error) {
	// Convert schema name to lowercase plural for the resource name
	plural := strings.ToLower(schema.Name) + "s"
	resourceName := plural + "." + opts.Group

	xrd := XRD{
		APIVersion: "apiextensions.crossplane.io/v1",
		Kind:       "CompositeResourceDefinition",
		Metadata: Metadata{
			Name: resourceName,
		},
		Spec: XRDSpec{
			Group: opts.Group,
			Names: Names{
				Kind:   schema.Name,
				Plural: plural,
			},
			Versions: []Version{
				{
					Name:          opts.Version,
					Served:        true,
					Referenceable: true,
					Schema: VersionSchema{
						OpenAPIV3Schema: OpenAPIV3Schema{
							Type:       "object",
							Properties: make(map[string]PropertySchema),
						},
					},
				},
			},
		},
	}

	// Add claim names if requested
	if opts.WithClaims {
		claimKind := opts.ClaimKind
		claimPlural := opts.ClaimPlural

		// Auto-generate claim names if not provided
		if claimKind == "" {
			// Remove 'X' prefix if present (common Crossplane convention)
			if strings.HasPrefix(schema.Name, "X") {
				claimKind = strings.TrimPrefix(schema.Name, "X")
			} else {
				claimKind = schema.Name
			}
		}

		if claimPlural == "" {
			claimPlural = strings.ToLower(claimKind) + "s"
		}

		xrd.Spec.ClaimNames = &ClaimNames{
			Kind:   claimKind,
			Plural: claimPlural,
		}
	}

	// Build the spec.parameters structure
	parametersSchema := PropertySchema{
		Type:       "object",
		Properties: make(map[string]PropertySchema),
		Required:   []string{},
	}

	for _, field := range schema.Fields {
		propSchema := convertFieldToPropertySchema(field)
		parametersSchema.Properties[field.Name] = propSchema
		if field.Required {
			parametersSchema.Required = append(parametersSchema.Required, field.Name)
		}
	}

	// Add spec section with parameters
	specSchema := PropertySchema{
		Type: "object",
		Properties: map[string]PropertySchema{
			"parameters": parametersSchema,
		},
		Required: []string{"parameters"},
	}

	xrd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"] = specSchema
	xrd.Spec.Versions[0].Schema.OpenAPIV3Schema.Required = []string{"spec"}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(xrd)
	if err != nil {
		return "", fmt.Errorf("failed to marshal XRD to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// convertFieldToPropertySchema converts a KCL field to an OpenAPI property schema
func convertFieldToPropertySchema(field parser.Field) PropertySchema {
	schema := PropertySchema{}

	// Map KCL types to OpenAPI types
	switch {
	case field.Type == "str":
		schema.Type = "string"
	case field.Type == "int":
		schema.Type = "integer"
	case field.Type == "float":
		schema.Type = "number"
	case field.Type == "bool":
		schema.Type = "boolean"
	case strings.HasPrefix(field.Type, "[") && strings.HasSuffix(field.Type, "]"):
		// Array type: [ElementType]
		schema.Type = "array"
		elementType := strings.TrimSuffix(strings.TrimPrefix(field.Type, "["), "]")
		elementSchema := convertFieldToPropertySchema(parser.Field{Type: elementType})
		schema.Items = &elementSchema
	case strings.HasPrefix(field.Type, "{") && strings.Contains(field.Type, ":"):
		// Map/dict type: {str: str}
		schema.Type = "object"
	default:
		// Assume it's an object type
		schema.Type = "object"
	}

	if field.Description != "" {
		schema.Description = field.Description
	}

	if field.Default != "" && field.Default != "Undefined" {
		// Parse the default value to remove quotes if it's a string literal
		defaultValue := strings.Trim(field.Default, `"`)
		
		// Try to convert to appropriate type
		switch schema.Type {
		case "integer":
			// Try to parse as integer
			if intVal, err := strconv.Atoi(defaultValue); err == nil {
				schema.Default = intVal
			} else {
				schema.Default = defaultValue
			}
		case "boolean":
			// Convert boolean strings to actual boolean values
			if defaultValue == "True" || defaultValue == "true" {
				schema.Default = true
			} else if defaultValue == "False" || defaultValue == "false" {
				schema.Default = false
			}
		case "number":
			// Try to parse as float
			if floatVal, err := strconv.ParseFloat(defaultValue, 64); err == nil {
				schema.Default = floatVal
			} else {
				schema.Default = defaultValue
			}
		case "string":
			schema.Default = defaultValue
		default:
			schema.Default = defaultValue
		}
	}

	return schema
}
