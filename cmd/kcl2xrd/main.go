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
	rootCmd.Flags().StringVarP(&group, "group", "g", "", "API group for the XRD (required)")
	rootCmd.Flags().StringVarP(&version, "version", "v", "v1alpha1", "API version for the XRD")
	rootCmd.Flags().StringVarP(&schemaName, "schema", "s", "", "Name of the schema to convert (defaults to last schema in file)")
	rootCmd.Flags().BoolVar(&withClaims, "with-claims", false, "Generate XRD with claimNames")
	rootCmd.Flags().StringVar(&claimKind, "claim-kind", "", "Kind for the claim (defaults to schema name without 'X' prefix)")
	rootCmd.Flags().StringVar(&claimPlural, "claim-plural", "", "Plural for the claim (auto-generated if not specified)")
	rootCmd.Flags().BoolVar(&served, "served", true, "Mark version as served")
	rootCmd.Flags().BoolVar(&referenceable, "referenceable", true, "Mark version as referenceable")
	rootCmd.Flags().StringSliceVar(&categories, "categories", nil, "Categories for the XRD (comma-separated)")
	rootCmd.Flags().StringSliceVar(&printerColumns, "printer-columns", nil, "Additional printer columns (format: name:type:jsonPath:description)")
	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("group")

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
		// User specified a schema name
		if result.Schemas[schemaName] == nil {
			return fmt.Errorf("schema '%s' not found in file. Available schemas: %v", schemaName, getSchemaNames(result.Schemas))
		}
		selectedSchema = result.Schemas[schemaName]
	} else {
		// Use the primary (last) schema
		selectedSchema = result.Primary
	}

	// Prepare generator options
	opts := generator.XRDOptions{
		Group:          group,
		Version:        version,
		WithClaims:     withClaims,
		ClaimKind:      claimKind,
		ClaimPlural:    claimPlural,
		Served:         served,
		Referenceable:  referenceable,
		Categories:     categories,
		PrinterColumns: parsePrinterColumns(printerColumns),
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
