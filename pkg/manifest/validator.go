package manifest

import (
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"

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
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension { //nolint:gocritic // singleCaseSwitch: this is intentional
	case constants.ManifestSuffixJsonnet:
		return v.validateJsonnet(filename)
	case constants.ManifestSuffixLibsonnet:
		return v.validateJsonnet(filename)
	}
	return nil
}

func (v *Validator) validateJsonnet(filename string) error {
	_, err := v.jsonnet.EvaluateFile(filename)
	return err
}
