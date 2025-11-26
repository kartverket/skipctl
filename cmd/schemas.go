package cmd

import (
	"fmt"

	"github.com/kartverket/skipctl/pkg/crd"
	"github.com/spf13/cobra"
)

var schemasCmd = &cobra.Command{
	Use:   "schemas",
	Short: "List supported Kubernetes schemas for validation",
	Long:  "Lists all Kubernetes CRD schemas that skipctl can validate against.",
	RunE:  runSchemas,
}

func runSchemas(_ *cobra.Command, _ []string) error {
	schemas, err := crd.ListSchemas()
	if err != nil {
		return err
	}

	fmt.Println("Supported schemas:")
	for _, schema := range schemas {
		fmt.Printf(" - %s\n", schema)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(schemasCmd)
}
