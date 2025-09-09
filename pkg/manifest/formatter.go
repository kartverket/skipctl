package manifest

import (
	"fmt"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
	"go.yaml.in/yaml/v4"
)

func formatJsonnet(file *Document) error {
	formatted, err := formatter.Format(file.Name, file.Content, formatter.DefaultOptions())
	if err != nil {
		return err
	}

	if werr := file.Write(formatted); werr != nil {
		return werr
	}

	return nil
}

func formatYaml(d *Document) error {
	var out any
	if err := yaml.Unmarshal([]byte(d.Content), &out); err != nil {
		return err
	}

	formattedYaml, merr := yaml.Marshal(out)
	if merr != nil {
		return merr
	}

	if werr := d.Write(string(formattedYaml)); werr != nil {
		return werr
	}

	return nil
}
func FormatManifest(file *Document) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		return formatJsonnet(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return formatYaml(file)
	default:
		return fmt.Errorf("invalid file format in file %s", file.Name)
	}
}
