package manifest

import (
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"

	"github.com/kartverket/skipctl/pkg/constants"
)

type Validator struct {
	k8s *K8sValidator
}

func NewValidator() *Validator {
	return &Validator{
		k8s: NewK8sValidator(),
	}
}

func (v *Validator) ValidateManifest(filename string) (ValidateResult, error) {
	extension := filepath.Ext(filename)

	switch extension {
	case constants.ManifestSuffixJsonnet:
		return v.validateJsonnet(filename)
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		return v.validateYaml(filename)
	}

	return ValidateResult{
		SkippedCount: 1,
	}, nil
}

func (v *Validator) validateJsonnet(filename string) (ValidateResult, error) {
	// There is a memory corruption bug that leads to segfaults if we reuse the same VM for multiple evaluations.
	// if there is a syntax error within the Jsonnet file, the VM gets corrupted and cannot be used again.
	vm := jsonnet.MakeVM()

	content, err := vm.EvaluateFile(filename)
	if err != nil {
		return ValidateResult{
			ErrorCount: 1,
		}, err
	}

	return v.k8s.validateK8sSchema(filename, content)
}

func (v *Validator) validateYaml(filename string) (ValidateResult, error) {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return ValidateResult{
			ErrorCount: 1,
		}, err
	}

	return v.k8s.validateK8sSchema(filename, string(fileContents))
}
