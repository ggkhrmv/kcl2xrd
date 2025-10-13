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
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       XRDSpec  `yaml:"spec"`
}

// Metadata represents the metadata section of an XRD
type Metadata struct {
	Name string `yaml:"name"`
}

// PrinterColumn represents an additional printer column
type PrinterColumn struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	JSONPath    string `yaml:"jsonPath"`
	Description string `yaml:"description,omitempty"`
	Priority    int    `yaml:"priority,omitempty"`
}

// XRDSpec represents the spec section of an XRD
type XRDSpec struct {
	Group      string      `yaml:"group"`
	Names      Names       `yaml:"names"`
	ClaimNames *ClaimNames `yaml:"claimNames,omitempty"`
	Categories []string    `yaml:"categories,omitempty"`
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
	Group          string
	Version        string
	WithClaims     bool
	ClaimKind      string
	ClaimPlural    string
	Served         bool
	Referenceable  bool
	Categories     []string
	PrinterColumns []PrinterColumn
}

// Version represents a version in an XRD spec
type Version struct {
	Name                   string          `yaml:"name"`
	Served                 bool            `yaml:"served"`
	Referenceable          bool            `yaml:"referenceable"`
	Schema                 VersionSchema   `yaml:"schema"`
	AdditionalPrinterColumns []PrinterColumn `yaml:"additionalPrinterColumns,omitempty"`
}

// VersionSchema represents the schema section of a version
type VersionSchema struct {
	OpenAPIV3Schema OpenAPIV3Schema `yaml:"openAPIV3Schema"`
}

// OpenAPIV3Schema represents an OpenAPI v3 schema
type OpenAPIV3Schema struct {
	Type       string                    `yaml:"type"`
	Properties map[string]PropertySchema `yaml:"properties,omitempty"`
	Required   []string                  `yaml:"required,omitempty"`
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
	// Validation fields
	Pattern                    string          `yaml:"pattern,omitempty"`
	MinLength                    *int            `yaml:"minLength,omitempty"`
	MaxLength                    *int            `yaml:"maxLength,omitempty"`
	Minimum                      *int            `yaml:"minimum,omitempty"`
	Maximum                      *int            `yaml:"maximum,omitempty"`
	Enum                         []string        `yaml:"enum,omitempty"`
	XKubernetesValidations       []K8sValidation `yaml:"x-kubernetes-validations,omitempty"`
	XKubernetesImmutable         *bool           `yaml:"x-kubernetes-immutable,omitempty"`
	XKubernetesPreserveUnknownFields *bool       `yaml:"x-kubernetes-preserve-unknown-fields,omitempty"`
	XKubernetesMapType           string          `yaml:"x-kubernetes-map-type,omitempty"`
	XKubernetesListType          string          `yaml:"x-kubernetes-list-type,omitempty"`
	XKubernetesListMapKeys       []string        `yaml:"x-kubernetes-list-map-keys,omitempty"`
}

// K8sValidation represents Kubernetes CEL validation rules
type K8sValidation struct {
	Rule    string `yaml:"rule"`
	Message string `yaml:"message,omitempty"`
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
	return GenerateXRDWithSchemasAndOptions(schema, nil, opts)
}

// GenerateXRDWithSchemasAndOptions generates a Crossplane XRD with schema resolution for nested types
func GenerateXRDWithSchemasAndOptions(schema *parser.Schema, schemas map[string]*parser.Schema, opts XRDOptions) (string, error) {
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
					Name:                     opts.Version,
					Served:                   opts.Served,
					Referenceable:            opts.Referenceable,
					AdditionalPrinterColumns: opts.PrinterColumns,
					Schema: VersionSchema{
						OpenAPIV3Schema: OpenAPIV3Schema{
							Type:       "object",
							Properties: make(map[string]PropertySchema),
						},
					},
				},
			},
			Categories: opts.Categories,
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
		propSchema := convertFieldToPropertySchemaWithSchemas(field, schemas)
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

	// Marshal to YAML with 2-space indentation
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	err := encoder.Encode(xrd)
	encoder.Close()
	if err != nil {
		return "", fmt.Errorf("failed to marshal XRD to YAML: %w", err)
	}

	return buf.String(), nil
}

// convertFieldToPropertySchema converts a KCL field to an OpenAPI property schema
func convertFieldToPropertySchema(field parser.Field) PropertySchema {
	return convertFieldToPropertySchemaWithSchemas(field, nil)
}

// convertFieldToPropertySchemaWithSchemas converts a KCL field to an OpenAPI property schema
// with support for nested schema expansion
func convertFieldToPropertySchemaWithSchemas(field parser.Field, schemas map[string]*parser.Schema) PropertySchema {
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
		
		// Check for [{any:any}] pattern - array of arbitrary objects
		if strings.TrimSpace(elementType) == "{any:any}" {
			// Array of objects with arbitrary properties
			elementSchema := PropertySchema{
				Type: "object",
			}
			// Apply preserve unknown fields if annotation is present
			if field.PreserveUnknownFields {
				preserve := true
				elementSchema.XKubernetesPreserveUnknownFields = &preserve
			}
			schema.Items = &elementSchema
		} else {
			elementSchema := convertFieldToPropertySchemaWithSchemas(parser.Field{Type: elementType}, schemas)
			schema.Items = &elementSchema
		}
	case strings.HasPrefix(field.Type, "{") && strings.Contains(field.Type, ":"):
		// Map/dict type: {str: str} or {any:any}
		schema.Type = "object"
		
		// Check for {any:any} pattern - object with arbitrary properties
		if strings.TrimSpace(strings.Trim(field.Type, "{}")) == "any:any" {
			// Object with arbitrary properties
			// Apply preserve unknown fields if annotation is present
			if field.PreserveUnknownFields {
				preserve := true
				schema.XKubernetesPreserveUnknownFields = &preserve
			}
		}
	default:
		// Check if it's a reference to another schema
		if schemas != nil && schemas[field.Type] != nil {
			// Expand the nested schema
			schema.Type = "object"
			schema.Properties = make(map[string]PropertySchema)
			nestedSchema := schemas[field.Type]
			
			// Add description from the field if present (for the object itself)
			if field.Description != "" {
				schema.Description = field.Description
			}
			
			for _, nestedField := range nestedSchema.Fields {
				nestedProp := convertFieldToPropertySchemaWithSchemas(nestedField, schemas)
				schema.Properties[nestedField.Name] = nestedProp
				if nestedField.Required {
					schema.Required = append(schema.Required, nestedField.Name)
				}
			}
			
			// Apply validation fields and defaults to the nested schema object
			applyFieldValidationsAndDefaults(field, &schema)
			
			// Return early since we've already handled description and validations
			return schema
		} else {
			// Assume it's an object type
			schema.Type = "object"
		}
	}

	if field.Description != "" {
		schema.Description = field.Description
	}
	
	applyFieldValidationsAndDefaults(field, &schema)

	return schema
}

// applyFieldValidationsAndDefaults applies validation and default values to a property schema
func applyFieldValidationsAndDefaults(field parser.Field, schema *PropertySchema) {

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
	
	// Apply validation constraints
	if field.Pattern != "" {
		schema.Pattern = field.Pattern
	}
	
	if field.MinLength != nil {
		schema.MinLength = field.MinLength
	}
	
	if field.MaxLength != nil {
		schema.MaxLength = field.MaxLength
	}
	
	if field.Minimum != nil {
		schema.Minimum = field.Minimum
	}
	
	if field.Maximum != nil {
		schema.Maximum = field.Maximum
	}
	
	if len(field.Enum) > 0 {
		schema.Enum = field.Enum
	}
	
	if field.Immutable {
		immutable := true
		schema.XKubernetesImmutable = &immutable
	}
	
	if field.PreserveUnknownFields {
		preserve := true
		schema.XKubernetesPreserveUnknownFields = &preserve
	}
	
	if field.MapType != "" {
		schema.XKubernetesMapType = field.MapType
	}
	
	if field.ListType != "" {
		schema.XKubernetesListType = field.ListType
	}
	
	if len(field.ListMapKeys) > 0 {
		schema.XKubernetesListMapKeys = field.ListMapKeys
	}
	
	// Apply CEL validations
	if len(field.CELValidations) > 0 {
		for _, celVal := range field.CELValidations {
			k8sVal := K8sValidation{
				Rule:    celVal.Rule,
				Message: celVal.Message,
			}
			schema.XKubernetesValidations = append(schema.XKubernetesValidations, k8sVal)
		}
	}
}
