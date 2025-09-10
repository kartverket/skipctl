package manifest

import (
	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/constants"
)

type Validator struct {
	k8s *K8sValidator
	res *ValidateResult
}

func NewValidator(tempDir string) *Validator {
	return &Validator{
		k8s: NewK8sValidator(tempDir),
		res: &ValidateResult{},
	}
}

var validateResult = &ValidateResult{}

func (v *Validator) ValidateManifest(file *Document) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		if jerr := v.handleValidateJsonnet(file); jerr != nil {
			return jerr
		}
		return nil
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		if yerr := v.handleValidateYaml(file); yerr != nil {
			return yerr
		}
		return nil
	}
	return nil
}
func (v *Validator) validateJsonnet(file *Document) (ValidateResult, error) {
	// There is a memory corruption bug that leads to segfaults if we reuse the same VM for multiple evaluations.
	// if there is a syntax error within the Jsonnet file, the VM gets corrupted and cannot be used again.
	vm := jsonnet.MakeVM()

	content, err := vm.EvaluateAnonymousSnippet(file.Name, file.Content)
	if err != nil {
		return ValidateResult{
			ErrorCount: 1,
		}, err
	}

	return v.k8s.validateK8sSchema(file.Name, content)
}

func (v *Validator) handleValidateJsonnet(d *Document) error {
	result, jerr := v.validateJsonnet(d)
	if jerr != nil {
		return jerr
	}
	countValidateRes(&result)
	return nil
}
func (v *Validator) validateYaml(file *Document) (ValidateResult, error) {
	return v.k8s.validateK8sSchema(file.Name, file.Content)
}
func (v *Validator) handleValidateYaml(d *Document) error {
	result, jerr := v.validateYaml(d)
	if jerr != nil {
		return jerr
	}
	countValidateRes(&result)
	return nil
}
func countValidateRes(result *ValidateResult) {
	validateResult.ErrorCount += result.ErrorCount
	validateResult.InvalidCount += result.InvalidCount
	validateResult.SkippedCount += result.SkippedCount
	validateResult.ValidCount += result.ValidCount
}
