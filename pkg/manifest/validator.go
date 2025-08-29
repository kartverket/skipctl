package manifest

import (
	"path/filepath"

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
	extension := filepath.Ext(filename)

	if extension == constants.Suffixes[0] {
		return v.validateJsonnet(filename)
	}

	return nil
}

func (v *Validator) validateJsonnet(filename string) error {
	_, err := v.jsonnet.EvaluateFile(filename)
	return err
}
