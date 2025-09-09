package manifest

import (
	"fmt"
	"io"
	"os"

	"github.com/google/go-jsonnet/formatter"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/utils"
	"go.yaml.in/yaml/v4"
)

func formatJsonnet(file *Document) error {
	formatted, err := formatter.Format(file.Name, file.Content, formatter.DefaultOptions())
	if err != nil {
		return err
	}

	if !file.FromStdin {
		fileInfo, fileInfoErr := os.Stat(file.Name)
		if fileInfoErr != nil {
			return fileInfoErr
		}
		if err = os.WriteFile(file.Name, []byte(formatted), fileInfo.Mode().Perm()); err != nil {
			return err
		}
	} else {
		_, err = io.WriteString(os.Stdout, formatted)
		if err != nil {
			return err
		}
	}

	return nil
}

func formatYaml(file *Document) error {
	if file.FromStdin {
		var out any
		if err := yaml.Unmarshal([]byte(file.Content), &out); err != nil {
			return err
		}
		formattedYaml, err := yaml.Marshal(out)
		if err != nil {
			return err
		}
		if _, err = os.Stdout.Write(formattedYaml); err != nil {
			return err
		}
		return nil
	}

	raw, err := utils.UnmarshalYamlFromFile(file.Name)
	if err != nil {
		return err
	}
	formattedYaml, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(file.Name)
	if err != nil {
		return err
	}
	if err = os.WriteFile(file.Name, formattedYaml, fileInfo.Mode().Perm()); err != nil {
		return err
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
