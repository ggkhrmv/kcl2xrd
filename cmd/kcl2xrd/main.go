package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ggkhrmv/kcl2xrd/pkg/generator"
	"github.com/ggkhrmv/kcl2xrd/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	inputFile          string
	outputFile         string
	group              string
	version            string
	withClaims         bool
	claimKind          string
	claimPlural        string
	schemaName         string
	served             bool
	referenceable      bool
	categories         []string
	printerColumns     []string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kcl2xrd",
		Short: "Convert KCL schemas to Crossplane XRDs",
		Long:  `A tool to convert KCL (KCL Configuration Language) schemas to Crossplane Composite Resource Definitions (XRDs)`,
		RunE:  run,
	}

	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input KCL schema file (required)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output XRD file (stdout if not specified)")
	rootCmd.Flags().StringVarP(&group, "group", "g", "", "API group for the XRD (optional if specified in KCL file via __xrd_group)")
	rootCmd.Flags().StringVarP(&version, "version", "v", "v1alpha1", "API version for the XRD")
	rootCmd.Flags().StringVarP(&schemaName, "schema", "s", "", "Name of the schema to convert (defaults to @xrd marked schema, __xrd_kind, or last schema in file)")
	rootCmd.Flags().BoolVar(&withClaims, "with-claims", false, "Generate XRD with claimNames")
	rootCmd.Flags().StringVar(&claimKind, "claim-kind", "", "Kind for the claim (defaults to schema name without 'X' prefix)")
	rootCmd.Flags().StringVar(&claimPlural, "claim-plural", "", "Plural for the claim (auto-generated if not specified)")
	rootCmd.Flags().BoolVar(&served, "served", true, "Mark version as served")
	rootCmd.Flags().BoolVar(&referenceable, "referenceable", true, "Mark version as referenceable")
	rootCmd.Flags().StringSliceVar(&categories, "categories", nil, "Categories for the XRD (comma-separated)")
	rootCmd.Flags().StringSliceVar(&printerColumns, "printer-columns", nil, "Additional printer columns (format: name:type:jsonPath:description)")
	
	if err := rootCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Parse KCL schema
	result, err := parser.ParseKCLFileWithSchemas(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse KCL file: %w", err)
	}

	// Select schema to convert
	var selectedSchema *parser.Schema
	if schemaName != "" {
		// User specified a schema name via CLI
		if result.Schemas[schemaName] == nil {
			return fmt.Errorf("schema '%s' not found in file. Available schemas: %v", schemaName, getSchemaNames(result.Schemas))
		}
		selectedSchema = result.Schemas[schemaName]
	} else {
		// Check if any schema is marked with @xrd annotation
		var xrdSchema *parser.Schema
		for _, schema := range result.Schemas {
			if schema.IsXRD {
				if xrdSchema != nil {
					return fmt.Errorf("multiple schemas marked with @xrd annotation: '%s' and '%s'. Only one schema should be marked.", xrdSchema.Name, schema.Name)
				}
				xrdSchema = schema
			}
		}
		
		if xrdSchema != nil {
			// Use the schema marked with @xrd
			selectedSchema = xrdSchema
		} else {
			// Use the primary (last) schema
			selectedSchema = result.Primary
		}
	}

	// Apply metadata from KCL file if present, CLI flags override
	if result.Metadata != nil {
		if group == "" && result.Metadata.Group != "" {
			group = result.Metadata.Group
		}
		if version == "v1alpha1" && result.Metadata.XRVersion != "" {
			version = result.Metadata.XRVersion
		}
		if len(categories) == 0 && len(result.Metadata.Categories) > 0 {
			categories = result.Metadata.Categories
		}
		if result.Metadata.Served != nil && !cmd.Flags().Changed("served") {
			served = *result.Metadata.Served
		}
		if result.Metadata.Referenceable != nil && !cmd.Flags().Changed("referenceable") {
			referenceable = *result.Metadata.Referenceable
		}
		if len(printerColumns) == 0 && len(result.Metadata.PrinterColumns) > 0 {
			// Convert parser.PrinterColumn to generator.PrinterColumn
			for _, pc := range result.Metadata.PrinterColumns {
				printerColumns = append(printerColumns, fmt.Sprintf("%s:%s:%s:%s", pc.Name, pc.Type, pc.JSONPath, pc.Description))
			}
		}
	}
	
	// Validate that group is provided
	if group == "" {
		return fmt.Errorf("API group must be specified either via --group flag or '__xrd_group' variable in KCL file")
	}

	// Prepare generator options
	opts := generator.XRDOptions{
		Group:          group,
		Version:        version,
		Kind:           "", // Will be set below if __xrd_kind is specified
		WithClaims:     withClaims,
		ClaimKind:      claimKind,
		ClaimPlural:    claimPlural,
		Served:         served,
		Referenceable:  referenceable,
		Categories:     categories,
		PrinterColumns: parsePrinterColumns(printerColumns),
	}
	
	// If __xrd_kind is specified in metadata, use it as the XRD kind
	if result.Metadata != nil && result.Metadata.XRKind != "" {
		opts.Kind = result.Metadata.XRKind
	}

	// Generate XRD with schema resolution
	xrd, err := generator.GenerateXRDWithSchemasAndOptions(selectedSchema, result.Schemas, opts)
	if err != nil {
		return fmt.Errorf("failed to generate XRD: %w", err)
	}

	// Output XRD
	if outputFile == "" {
		fmt.Println(xrd)
	} else {
		if err := os.WriteFile(outputFile, []byte(xrd), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "XRD written to %s\n", outputFile)
	}

	return nil
}

func getSchemaNames(schemas map[string]*parser.Schema) []string {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	return names
}

func parsePrinterColumns(columns []string) []generator.PrinterColumn {
	result := make([]generator.PrinterColumn, 0, len(columns))
	for _, col := range columns {
		parts := strings.Split(col, ":")
		if len(parts) >= 3 {
			pc := generator.PrinterColumn{
				Name:     parts[0],
				Type:     parts[1],
				JSONPath: parts[2],
			}
			if len(parts) >= 4 {
				pc.Description = parts[3]
			}
			result = append(result, pc)
		}
	}
	return result
}
