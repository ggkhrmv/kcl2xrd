package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	kcl "kcl-lang.io/kcl-go"
)

// Schema represents a parsed KCL schema
type Schema struct {
	Name        string
	Description string
	Fields      []Field
	IsXRD       bool // marked with @xrd annotation
}

// ParseResult contains all schemas parsed from a file
type ParseResult struct {
	Schemas  map[string]*Schema // map of schema name to schema
	Primary  *Schema            // the last/main schema in the file
	Metadata *XRDMetadata       // XRD metadata from KCL variables
}

// XRDMetadata contains metadata for XRD generation parsed from KCL variables
type XRDMetadata struct {
	XRKind         string
	XRVersion      string
	Group          string
	Categories     []string
	PrinterColumns []PrinterColumn
	Served         *bool
	Referenceable  *bool
}

// PrinterColumn represents an additional printer column
type PrinterColumn struct {
	Name        string
	Type        string
	JSONPath    string
	Description string
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
	MinItems    *int   // minimum number of items in arrays
	Enum        []string // enumeration of allowed values
	Immutable   bool   // x-kubernetes-immutable marker
	CELValidations []CELValidation // CEL validation rules
	// Kubernetes-specific annotations
	PreserveUnknownFields bool   // x-kubernetes-preserve-unknown-fields
	MapType               string // x-kubernetes-map-type
	ListType              string // x-kubernetes-list-type
	ListMapKeys           []string // x-kubernetes-list-map-keys
}

// CELValidation represents a CEL validation rule
type CELValidation struct {
	Rule    string
	Message string
}

// ParseKCLFile parses a KCL schema file and returns a Schema structure
// For backward compatibility, it returns the primary (last) schema
func ParseKCLFile(filename string) (*Schema, error) {
	result, err := ParseKCLFileWithSchemas(filename)
	if err != nil {
		return nil, err
	}
	return result.Primary, nil
}

// ParseKCLFileWithSchemas parses a KCL schema file and returns all schemas
func ParseKCLFileWithSchemas(filename string) (*ParseResult, error) {
	// First, try to evaluate metadata using KCL runtime for more flexibility
	kclMetadata, _ := evaluateMetadataWithKCL(filename)
	
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSchema *Schema
	var currentField *Field
	var inSchema bool
	var inDocstring bool
	var docstringLines []string
	
	schemas := make(map[string]*Schema)
	var primarySchema *Schema
	metadata := &XRDMetadata{}

	schemaRegex := regexp.MustCompile(`^\s*schema\s+(\w+)\s*:?\s*$`)
	fieldRegex := regexp.MustCompile(`^\s*(\w+)\s*(\?)?:\s*(.+?)(?:\s*=\s*(.+))?\s*$`)
	
	// Metadata variable patterns (using __xrd_ prefix for unique naming)
	xrKindRegex := regexp.MustCompile(`^\s*__xrd_kind\s*=\s*['"](.*?)['"]\s*$`)
	xrVersionRegex := regexp.MustCompile(`^\s*__xrd_version\s*=\s*['"](.*?)['"]\s*$`)
	groupRegex := regexp.MustCompile(`^\s*__xrd_group\s*=\s*['"](.*?)['"]\s*$`)
	// Also match __xrd_group with any expression (skip parsing, user must provide via CLI)
	groupExprRegex := regexp.MustCompile(`^\s*__xrd_group\s*=\s*(.+)$`)
	categoriesRegex := regexp.MustCompile(`^\s*__xrd_categories\s*=\s*\[(.*?)\]\s*$`)
	servedRegex := regexp.MustCompile(`^\s*__xrd_served\s*=\s*(true|false|True|False)\s*$`)
	referenceableRegex := regexp.MustCompile(`^\s*__xrd_referenceable\s*=\s*(true|false|True|False)\s*$`)
	printerColumnsRegex := regexp.MustCompile(`^\s*__xrd_printer_columns\s*=\s*\[(.*?)\]\s*$`)
	
	// Validation annotation patterns
	patternRegex := regexp.MustCompile(`@pattern\s*\(\s*['"](.*?)['"]\s*\)`)
	minLengthRegex := regexp.MustCompile(`@minLength\s*\(\s*(\d+)\s*\)`)
	maxLengthRegex := regexp.MustCompile(`@maxLength\s*\(\s*(\d+)\s*\)`)
	minimumRegex := regexp.MustCompile(`@minimum\s*\(\s*(\d+)\s*\)`)
	maximumRegex := regexp.MustCompile(`@maximum\s*\(\s*(\d+)\s*\)`)
	minItemsRegex := regexp.MustCompile(`@minItems\s*\(\s*(\d+)\s*\)`)
	enumRegex := regexp.MustCompile(`@enum\s*\(\s*\[(.*?)\]\s*\)`)
	immutableRegex := regexp.MustCompile(`@immutable`)
	celValidationRegex := regexp.MustCompile(`@validate\s*\(\s*['"](.*?)['"]\s*(?:,\s*['"](.*?)['"]\s*)?\)`)
	preserveUnknownFieldsRegex := regexp.MustCompile(`@preserveUnknownFields`)
	mapTypeRegex := regexp.MustCompile(`@mapType\s*\(\s*['"](.*?)['"]\s*\)`)
	listTypeRegex := regexp.MustCompile(`@listType\s*\(\s*['"](.*?)['"]\s*\)`)
	listMapKeysRegex := regexp.MustCompile(`@listMapKeys\s*\(\s*\[(.*?)\]\s*\)`)
	xrdAnnotationRegex := regexp.MustCompile(`@xrd`)
	
	var pendingAnnotations []string
	var pendingComments []string
	
	// Track variable assignments for resolving expressions
	variables := make(map[string]string)
	// Regex for simple variable assignments like: _xrSubgroup = "aws"
	varAssignRegex := regexp.MustCompile(`^\s*(_\w+)\s*=\s*['"](.*?)['"]\s*$`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines (except docstrings and when collecting comments)
		if trimmedLine == "" && !inDocstring {
			// Clear pending comments on empty line if not in schema
			if !inSchema {
				pendingComments = nil
			}
			continue
		}
		
		// Parse metadata variables (before schema definitions)
		if !inSchema {
			// Track variable assignments for later resolution
			if matches := varAssignRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				varName := matches[1]
				varValue := matches[2]
				variables[varName] = varValue
			}
			
			if matches := xrKindRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				metadata.XRKind = matches[1]
				continue
			}
			if matches := xrVersionRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				metadata.XRVersion = matches[1]
				continue
			}
			if matches := groupRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				metadata.Group = matches[1]
				continue
			}
			// If __xrd_group doesn't match the simple pattern, try to resolve format expressions
			if groupExprRegex.MatchString(trimmedLine) && !groupRegex.MatchString(trimmedLine) {
				// Try to resolve format expressions like: "{}.{}".format(var1, var2)
				if resolvedGroup := resolveFormatExpression(trimmedLine, variables); resolvedGroup != "" {
					metadata.Group = resolvedGroup
				}
				// If resolution failed, user will need to provide --group flag
				continue
			}
			if matches := categoriesRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				categoriesStr := matches[1]
				categories := strings.Split(categoriesStr, ",")
				for i, cat := range categories {
					cat = strings.TrimSpace(cat)
					cat = strings.Trim(cat, `"'`)
					categories[i] = cat
				}
				metadata.Categories = categories
				continue
			}
			if matches := servedRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				served := strings.ToLower(matches[1]) == "true"
				metadata.Served = &served
				continue
			}
			if matches := referenceableRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				referenceable := strings.ToLower(matches[1]) == "true"
				metadata.Referenceable = &referenceable
				continue
			}
			if matches := printerColumnsRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				columnsStr := matches[1]
				// Parse printer columns format: "Name:string:.metadata.name:Description", "Age:integer:.status.age:Age in days"
				columnStrs := splitPrinterColumns(columnsStr)
				for _, colStr := range columnStrs {
					parts := strings.Split(colStr, ":")
					if len(parts) >= 3 {
						pc := PrinterColumn{
							Name:     parts[0],
							Type:     parts[1],
							JSONPath: parts[2],
						}
						if len(parts) >= 4 {
							pc.Description = parts[3]
						}
						metadata.PrinterColumns = append(metadata.PrinterColumns, pc)
					}
				}
				continue
			}
		}
		
		// Check for comments (annotations and descriptions)
		if strings.HasPrefix(trimmedLine, "#") && !inDocstring {
			commentText := strings.TrimPrefix(trimmedLine, "#")
			commentText = strings.TrimSpace(commentText)
			
			// Check if it's an annotation
			if strings.HasPrefix(commentText, "@") {
				pendingAnnotations = append(pendingAnnotations, trimmedLine)
			} else if inSchema {
				// It's a regular comment - store for next field description
				pendingComments = append(pendingComments, commentText)
			}
			continue
		}

		// Handle docstrings
		if strings.Contains(trimmedLine, `"""`) || strings.Contains(trimmedLine, `r"""`) {
			if inDocstring {
				inDocstring = false
				if currentField != nil {
					currentField.Description = strings.Join(docstringLines, " ")
				} else if currentSchema != nil && currentSchema.Description == "" {
					currentSchema.Description = strings.Join(docstringLines, " ")
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
			// Save previous schema if exists
			if currentSchema != nil {
				schemas[currentSchema.Name] = currentSchema
				primarySchema = currentSchema
			}
			
			// Start new schema
			currentSchema = &Schema{
				Name:   matches[1],
				Fields: []Field{},
			}
			
			// Check if this schema is marked with @xrd annotation
			for _, annotation := range pendingAnnotations {
				if xrdAnnotationRegex.MatchString(annotation) {
					currentSchema.IsXRD = true
					break
				}
			}
			pendingAnnotations = nil
			pendingComments = nil
			
			inSchema = true
			continue
		}

		// Check if we should exit schema (non-indented line that's not empty or comment)
		if inSchema && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmedLine != "" {
			// Non-indented line - schema has ended
			inSchema = false
			currentField = nil
		}
		
		// Parse field definitions
		if inSchema && currentSchema != nil {
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
				
				// Remove inline comments from default value if present
				if defaultValue != "" && strings.Contains(defaultValue, "#") {
					parts := strings.SplitN(defaultValue, "#", 2)
					defaultValue = strings.TrimSpace(parts[0])
				}

				field := Field{
					Name:     fieldName,
					Type:     fieldType,
					Required: !optional,
					Default:  defaultValue,
				}
				
				// Set description from pending comments (above field)
				if len(pendingComments) > 0 {
					field.Description = strings.Join(pendingComments, "\n")
					pendingComments = nil
				}
				
				// Apply validation annotations from pending comments
				applyValidationAnnotations(&field, pendingAnnotations, 
					patternRegex, minLengthRegex, maxLengthRegex, 
					minimumRegex, maximumRegex, minItemsRegex, enumRegex, immutableRegex, celValidationRegex,
					preserveUnknownFieldsRegex, mapTypeRegex, listTypeRegex, listMapKeysRegex)
				pendingAnnotations = nil
				
				currentSchema.Fields = append(currentSchema.Fields, field)
				currentField = &currentSchema.Fields[len(currentSchema.Fields)-1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	
	// Save the last schema
	if currentSchema != nil {
		schemas[currentSchema.Name] = currentSchema
		primarySchema = currentSchema
	}

	if primarySchema == nil {
		return nil, fmt.Errorf("no schema found in file")
	}
	
	// Merge KCL-evaluated metadata with manually parsed metadata
	// KCL evaluation takes priority as it's more accurate
	if kclMetadata.XRKind != "" {
		metadata.XRKind = kclMetadata.XRKind
	}
	if kclMetadata.Group != "" {
		metadata.Group = kclMetadata.Group
	}
	if kclMetadata.XRVersion != "" {
		metadata.XRVersion = kclMetadata.XRVersion
	}
	if len(kclMetadata.Categories) > 0 {
		metadata.Categories = kclMetadata.Categories
	}
	if kclMetadata.Served != nil {
		metadata.Served = kclMetadata.Served
	}
	if kclMetadata.Referenceable != nil {
		metadata.Referenceable = kclMetadata.Referenceable
	}

	return &ParseResult{
		Schemas:  schemas,
		Primary:  primarySchema,
		Metadata: metadata,
	}, nil
}

// applyValidationAnnotations applies validation annotations from comments to a field
func applyValidationAnnotations(field *Field, annotations []string, 
	patternRegex, minLengthRegex, maxLengthRegex, minimumRegex, maximumRegex, minItemsRegex, enumRegex, immutableRegex, celValidationRegex,
	preserveUnknownFieldsRegex, mapTypeRegex, listTypeRegex, listMapKeysRegex *regexp.Regexp) {
	
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
		
		// Check for minItems
		if matches := minItemsRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				field.MinItems = &val
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
		
		// Check for preserveUnknownFields
		if preserveUnknownFieldsRegex.MatchString(annotation) {
			field.PreserveUnknownFields = true
		}
		
		// Check for mapType
		if matches := mapTypeRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			field.MapType = matches[1]
		}
		
		// Check for listType
		if matches := listTypeRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			field.ListType = matches[1]
		}
		
		// Check for listMapKeys
		if matches := listMapKeysRegex.FindStringSubmatch(annotation); len(matches) > 1 {
			keysStr := matches[1]
			keys := strings.Split(keysStr, ",")
			for i, key := range keys {
				key = strings.TrimSpace(key)
				key = strings.Trim(key, `"'`)
				keys[i] = key
			}
			field.ListMapKeys = keys
		}
	}
}

// splitPrinterColumns splits printer columns string respecting quoted strings
func splitPrinterColumns(s string) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)
	
	for i, ch := range s {
		if (ch == '"' || ch == '\'') && (i == 0 || s[i-1] != '\\') {
			if inQuote {
				if ch == quoteChar {
					inQuote = false
					quoteChar = 0
				}
			} else {
				inQuote = true
				quoteChar = ch
			}
			continue
		}
		
		if ch == ',' && !inQuote {
			trimmed := strings.TrimSpace(current.String())
			trimmed = strings.Trim(trimmed, `"'`)
			if trimmed != "" {
				result = append(result, trimmed)
			}
			current.Reset()
			continue
		}
		
		current.WriteRune(ch)
	}
	
	// Add last item
	trimmed := strings.TrimSpace(current.String())
	trimmed = strings.Trim(trimmed, `"'`)
	if trimmed != "" {
		result = append(result, trimmed)
	}
	
	return result
}

// resolveFormatExpression attempts to resolve KCL format expressions like:
// __xrd_group = "{}.{}".format(_xrSubgroup, _platformGroup)
// Returns the resolved string if successful, empty string otherwise
func resolveFormatExpression(line string, variables map[string]string) string {
	// Pattern to match: __xrd_group = "format_string".format(var1, var2, ...)
	formatRegex := regexp.MustCompile(`^\s*__xrd_group\s*=\s*["'](.*?)["']\.format\((.*?)\)\s*$`)
	matches := formatRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return ""
	}
	
	formatStr := matches[1]
	argsStr := matches[2]
	
	// Parse the arguments
	args := strings.Split(argsStr, ",")
	var resolvedArgs []string
	
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		// Remove leading/trailing quotes if present
		arg = strings.Trim(arg, `"'`)
		
		// Look up variable value
		if val, exists := variables[arg]; exists {
			resolvedArgs = append(resolvedArgs, val)
		} else {
			// Variable not found - cannot resolve this expression
			// This includes cases like settings.PLATFORM_API_GROUP which aren't simple variables
			return ""
		}
	}
	
	// Replace {} placeholders with actual values
	result := formatStr
	for _, val := range resolvedArgs {
		result = strings.Replace(result, "{}", val, 1)
	}
	
	// Check if all placeholders were replaced
	if strings.Contains(result, "{}") {
		// Still has unreplaced placeholders
		return ""
	}
	
	return result
}

// evaluateMetadataWithKCL uses KCL runtime to evaluate metadata variables
// This is more flexible than parsing format strings manually
func evaluateMetadataWithKCL(filename string) (*XRDMetadata, error) {
	metadata := &XRDMetadata{}
	
	// First, try to run KCL with the file as-is (with imports)
	// This allows imports to work when they can be resolved
	result, err := kcl.RunFiles([]string{filename}, kcl.WithShowHidden(true))
	if err != nil {
		// If evaluation failed (possibly due to unresolvable imports),
		// try again with imports filtered out
		content, readErr := os.ReadFile(filename)
		if readErr != nil {
			return metadata, nil
		}
		
		// Filter out import statements
		lines := strings.Split(string(content), "\n")
		var filteredLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Skip import statements
			if strings.HasPrefix(trimmed, "import ") {
				continue
			}
			filteredLines = append(filteredLines, line)
		}
		filteredContent := strings.Join(filteredLines, "\n")
		
		// Try running without imports
		result, err = kcl.Run("", kcl.WithCode(filteredContent), kcl.WithShowHidden(true))
		if err != nil {
			// If it still fails, return empty metadata (will fall back to manual parsing)
			return metadata, nil
		}
	}
	
	// Extract metadata variables from the result
	kclResult := result.First()
	if kclResult == nil {
		return metadata, nil
	}
	
	// Convert to map
	resultMap, err := kclResult.ToMap()
	if err != nil {
		return metadata, nil
	}
	
	// Try to extract __xrd_kind
	if kind, ok := resultMap["__xrd_kind"].(string); ok {
		metadata.XRKind = kind
	}
	
	// Try to extract __xrd_group
	if group, ok := resultMap["__xrd_group"].(string); ok {
		metadata.Group = group
	}
	
	// Try to extract __xrd_version
	if version, ok := resultMap["__xrd_version"].(string); ok {
		metadata.XRVersion = version
	}
	
	// Try to extract __xrd_served
	if served, ok := resultMap["__xrd_served"].(bool); ok {
		metadata.Served = &served
	}
	
	// Try to extract __xrd_referenceable
	if referenceable, ok := resultMap["__xrd_referenceable"].(bool); ok {
		metadata.Referenceable = &referenceable
	}
	
	// Try to extract __xrd_categories
	if categories, ok := resultMap["__xrd_categories"].([]interface{}); ok {
		for _, cat := range categories {
			if catStr, ok := cat.(string); ok {
				metadata.Categories = append(metadata.Categories, catStr)
			}
		}
	}
	
	return metadata, nil
}
