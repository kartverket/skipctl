package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
	fileutil "github.com/kartverket/skipctl/pkg/util"
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
	raw, err := fileutil.UnmarshalYamlFromFile(filename)
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
	switch filepath.Ext(strings.ToLower(filename)) {
	case constants.ManifestSuffixJsonnet:
		return formatJsonnet(filename)
	case constants.ManifestSuffixYaml:
		return formatYaml(filename)
	case constants.ManifestSuffixYml:
		return formatYaml(filename)
	default:
		return fmt.Errorf("invalid file format in file %s", filename)
	}
}
