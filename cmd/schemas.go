package cmd

import (
	"fmt"
	"strings"

	"github.com/kartverket/skipctl/pkg/crd"
	"github.com/kartverket/skipctl/pkg/logging"
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

	var sb strings.Builder
	sb.WriteString("Supported schemas:\n")
	for _, schema := range schemas {
		sb.WriteString(fmt.Sprintf(" - %s\n", schema))
	}
	logging.RawLogger().Info(sb.String())
	return nil
}

func init() {
	rootCmd.AddCommand(schemasCmd)
}
