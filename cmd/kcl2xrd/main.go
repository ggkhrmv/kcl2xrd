package main

import (
	"fmt"
	"os"

	"github.com/ggkhrmv/kcl2xrd/pkg/generator"
	"github.com/ggkhrmv/kcl2xrd/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	inputFile  string
	outputFile string
	group      string
	version    string
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
	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("group")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Parse KCL schema
	schema, err := parser.ParseKCLFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse KCL file: %w", err)
	}

	// Generate XRD
	xrd, err := generator.GenerateXRD(schema, group, version)
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
