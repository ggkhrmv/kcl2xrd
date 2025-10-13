package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Schema represents a parsed KCL schema
type Schema struct {
	Name        string
	Description string
	Fields      []Field
}

// Field represents a field in a KCL schema
type Field struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     string
	// Validation fields
	Pattern     string // regex pattern for string validation
	MinLength   *int   // minimum length for strings
	MaxLength   *int   // maximum length for strings
	Minimum     *int   // minimum value for numbers
	Maximum     *int   // maximum value for numbers
	Enum        []string // enumeration of allowed values
	Immutable   bool   // x-kubernetes-immutable marker
	CELValidations []CELValidation // CEL validation rules
}

// CELValidation represents a CEL validation rule
type CELValidation struct {
	Rule    string
	Message string
}

// ParseKCLFile parses a KCL schema file and returns a Schema structure
func ParseKCLFile(filename string) (*Schema, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var schema *Schema
	var currentField *Field
	var inSchema bool
	var inDocstring bool
	var docstringLines []string

	schemaRegex := regexp.MustCompile(`^\s*schema\s+(\w+)\s*:?\s*$`)
	fieldRegex := regexp.MustCompile(`^\s*(\w+)\s*(\?)?:\s*(.+?)(?:\s*=\s*(.+))?\s*$`)
	
	// Validation annotation patterns
	patternRegex := regexp.MustCompile(`@pattern\s*\(\s*['"](.*?)['"]\s*\)`)
	minLengthRegex := regexp.MustCompile(`@minLength\s*\(\s*(\d+)\s*\)`)
	maxLengthRegex := regexp.MustCompile(`@maxLength\s*\(\s*(\d+)\s*\)`)
	minimumRegex := regexp.MustCompile(`@minimum\s*\(\s*(\d+)\s*\)`)
	maximumRegex := regexp.MustCompile(`@maximum\s*\(\s*(\d+)\s*\)`)
	enumRegex := regexp.MustCompile(`@enum\s*\(\s*\[(.*?)\]\s*\)`)
	immutableRegex := regexp.MustCompile(`@immutable`)
	celValidationRegex := regexp.MustCompile(`@validate\s*\(\s*['"](.*?)['"]\s*(?:,\s*['"](.*?)['"]\s*)?\)`)
	
	var pendingAnnotations []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments (except docstrings)
		if trimmedLine == "" && !inDocstring {
			continue
		}
		
		// Check for validation annotations in comments
		if strings.HasPrefix(trimmedLine, "#") && !inDocstring {
			// Store annotation for next field
			pendingAnnotations = append(pendingAnnotations, trimmedLine)
			continue
		}

		// Handle docstrings
		if strings.Contains(trimmedLine, `"""`) || strings.Contains(trimmedLine, `r"""`) {
			if inDocstring {
				inDocstring = false
				if currentField != nil {
					currentField.Description = strings.Join(docstringLines, " ")
				} else if schema != nil && schema.Description == "" {
					schema.Description = strings.Join(docstringLines, " ")
				}
				docstringLines = nil
			} else {
				inDocstring = true
			}
			continue
		}

		if inDocstring {
			docstringLines = append(docstringLines, trimmedLine)
			continue
		}

		// Check for schema definition
		if matches := schemaRegex.FindStringSubmatch(line); matches != nil {
			if schema != nil {
				// We found another schema, we only parse the first one
				break
			}
			schema = &Schema{
				Name:   matches[1],
				Fields: []Field{},
			}
			inSchema = true
			continue
		}

		// Parse field definitions
		if inSchema && schema != nil {
			if matches := fieldRegex.FindStringSubmatch(line); matches != nil {
				fieldName := matches[1]
				optional := matches[2] == "?"
				fieldType := strings.TrimSpace(matches[3])
				defaultValue := ""
				if len(matches) > 4 {
					defaultValue = strings.TrimSpace(matches[4])
				}

				// Clean up the type (remove "default is" text if present)
				if strings.Contains(fieldType, ",") {
					parts := strings.Split(fieldType, ",")
					fieldType = strings.TrimSpace(parts[0])
				}

				field := Field{
					Name:     fieldName,
					Type:     fieldType,
					Required: !optional,
					Default:  defaultValue,
				}
				
				// Apply validation annotations from pending comments
				applyValidationAnnotations(&field, pendingAnnotations, 
					patternRegex, minLengthRegex, maxLengthRegex, 
					minimumRegex, maximumRegex, enumRegex, immutableRegex, celValidationRegex)
				pendingAnnotations = nil
				
				schema.Fields = append(schema.Fields, field)
				currentField = &schema.Fields[len(schema.Fields)-1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if schema == nil {
		return nil, fmt.Errorf("no schema found in file")
	}

	return schema, nil
}

// applyValidationAnnotations applies validation annotations from comments to a field
func applyValidationAnnotations(field *Field, annotations []string, 
	patternRegex, minLengthRegex, maxLengthRegex, minimumRegex, maximumRegex, enumRegex, immutableRegex, celValidationRegex *regexp.Regexp) {
	
	for _, annotation := range annotations {
		// Check for pattern
		if matches := patternRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			field.Pattern = matches[1]
		}
		
		// Check for minLength
		if matches := minLengthRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				field.MinLength = &val
			}
		}
		
		// Check for maxLength
		if matches := maxLengthRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				field.MaxLength = &val
			}
		}
		
		// Check for minimum
		if matches := minimumRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				field.Minimum = &val
			}
		}
		
		// Check for maximum
		if matches := maximumRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				field.Maximum = &val
			}
		}
		
		// Check for enum
		if matches := enumRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			enumStr := matches[1]
			// Split by comma and clean up
			enumValues := strings.Split(enumStr, ",")
			for i, val := range enumValues {
				val = strings.TrimSpace(val)
				val = strings.Trim(val, `"'`)
				enumValues[i] = val
			}
			field.Enum = enumValues
		}
		
		// Check for immutable
		if immutableRegex.MatchString(annotation) {
			field.Immutable = true
		}
		
		// Check for CEL validation
		if matches := celValidationRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			rule := matches[1]
			message := ""
			if len(matches) > 2 && matches[2] != "" {
				message = matches[2]
			}
			field.CELValidations = append(field.CELValidations, CELValidation{
				Rule:    rule,
				Message: message,
			})
		}
	}
}
