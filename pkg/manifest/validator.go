package manifest

import (
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"go.yaml.in/yaml/v4"

	"github.com/kartverket/skipctl/pkg/constants"
)

type Validator struct {
	jsonnet *jsonnet.VM
}

func NewValidator() *Validator {
	return &Validator{
		jsonnet: jsonnet.MakeVM(),
	}
}

func (v *Validator) ValidateManifest(filename string) error {
	extension := filepath.Ext(filename)

	switch extension {
	case constants.ManifestSuffixJsonnet:
		return v.validateJsonnet(filename)
	case constants.ManifestSuffixYaml:
		return v.validateYaml(filename)
	}
	return nil
}

func (v *Validator) validateJsonnet(filename string) error {
	_, err := v.jsonnet.EvaluateFile(filename)
	return err
}

func (v *Validator) validateYaml(filename string) error {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var output any

	return yaml.Unmarshal(fileContents, &output)
}
