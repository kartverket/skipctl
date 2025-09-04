package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/utils"
	"go.yaml.in/yaml/v4"
)

func formatJsonnet(filename string) error {
	rawContents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	formatted, err := formatter.Format(filename, string(rawContents), formatter.DefaultOptions())
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filename, []byte(formatted), fileInfo.Mode().Perm()); err != nil {
		return err
	}

	return nil
}

func formatYaml(filename string) error {
	raw, err := utils.UnmarshalYamlFromFile(filename)
	if err != nil {
		return err
	}

	formattedYaml, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filename, formattedYaml, fileInfo.Mode().Perm()); err != nil {
		return err
	}
	return nil
}
func FormatManifest(filename string) error {
	switch strings.ToLower(filepath.Ext(filename)) {
	case constants.ManifestSuffixJsonnet:
		return formatJsonnet(filename)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return formatYaml(filename)
	default:
		return fmt.Errorf("invalid file format in file %s", filename)
	}
}
