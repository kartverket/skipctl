package format

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/manifest"
	"go.yaml.in/yaml/v4"
)

func formatJsonnet(file *manifest.Document) error {
	formatted, err := formatter.Format(file.Name, file.Content, formatter.DefaultOptions())
	if err != nil {
		return err
	}

	if werr := file.Write(formatted); werr != nil {
		return werr
	}

	return nil
}

func formatYaml(d *manifest.Document) error {
	decoder := yaml.NewDecoder(strings.NewReader(d.Content))
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(constants.YamlIndent)

	for {
		var node yaml.Node
		decErr := decoder.Decode(&node)
		if decErr != nil {
			if errors.Is(decErr, io.EOF) {
				break
			}
			return fmt.Errorf("failed to decode YAML: %w", decErr)
		}

		if len(node.Content) == 0 {
			continue
		}

		if encErr := encoder.Encode(&node); encErr != nil {
			return fmt.Errorf("failed to encode YAML: %w", encErr)
		}
	}

	if werr := d.Write(output.String()); werr != nil {
		return werr
	}
	return nil
}

func Manifest(file *manifest.Document) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet, constants.ManifestSuffixLibsonnet:
		return formatJsonnet(file)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return formatYaml(file)
	default:
		return fmt.Errorf("invalid file format in file %s", file.Name)
	}
}
