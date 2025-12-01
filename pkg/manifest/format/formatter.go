package format

import (
	"fmt"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"go.yaml.in/yaml/v4"
)

func Jsonnet(file *manifest.Document) error {
	formatted, err := formatter.Format(file.Name, file.Content, formatter.DefaultOptions())
	if err != nil {
		return err
	}

	if werr := file.Write(formatted); werr != nil {
		return werr
	}

	return nil
}

func Yaml(d *manifest.Document) error {
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

func Manifest(file *manifest.Document) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet, constants.ManifestSuffixLibsonnet:
		return Jsonnet(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return Yaml(file)
	default:
		return fmt.Errorf("invalid file format in file %s", file.Name)
	}
}
