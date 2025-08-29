package manifest

import (
	"path/filepath"

	"github.com/google/go-jsonnet"

	"github.com/kartverket/skipctl/pkg/constants"
)

type ManifestValidator struct {
	jsonnet *jsonnet.VM
}

func NewManifestValidator() *ManifestValidator {
	return &ManifestValidator{
		jsonnet: jsonnet.MakeVM(),
	}
}

func (v *ManifestValidator) ValidateManifest(filename string) error {

	switch filepath.Ext(filename) {
	case constants.Suffixes[0]:
		return v.validateJsonnet(filename)

	}
	return nil

}

func (v *ManifestValidator) validateJsonnet(filename string) error {
	_, err := v.jsonnet.EvaluateFile(filename)
	return err
}
