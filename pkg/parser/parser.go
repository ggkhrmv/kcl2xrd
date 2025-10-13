package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments (except docstrings)
		if trimmedLine == "" && !inDocstring {
			continue
		}
		if strings.HasPrefix(trimmedLine, "#") && !inDocstring {
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
