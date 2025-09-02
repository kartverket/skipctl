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
	k8s     *K8sValidator
}

func NewValidator() *Validator {
	return &Validator{
		k8s: NewK8sValidator(),
	}
}

func (v *Validator) ValidateManifest(filename string) (ValidateResult, error) {
	extension := filepath.Ext(filename)

	switch extension { //nolint:gocritic // singleCaseSwitch: this is intentional
	case constants.ManifestSuffixJsonnet:
		return v.validateJsonnet(filename)
	}

	return ValidateResult{
		ValidCount:   0,
		InvalidCount: 0,
		ErrorCount:   0,
		SkippedCount: 1,
	}, nil
}

func (v *Validator) validateJsonnet(filename string) (result ValidateResult, err error) {

	// There is a memory corruption bug that leads to segfaults if we reuse the same VM for multiple evaluations.
	// if there is a syntax error within the Jsonnet file, the VM gets corrupted and cannot be used again.
	vm := jsonnet.MakeVM()

	content, err := vm.EvaluateFile(filename)
	if err != nil {
		return ValidateResult{
			ValidCount:   0,
			InvalidCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
		}, err
	}

	return v.k8s.validateK8sSchema(filename, content)
}

func (v *Validator) validateYaml(filename string) (ValidateResult, error) {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return ValidateResult{
			ValidCount:   0,
			InvalidCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
		}, err
	}

	summary, err := v.k8s.validateK8sSchema(filename, string(fileContents))
	if err != nil {
		return summary, err
	}

	var output any

	return summary, yaml.Unmarshal(fileContents, &output)
}
