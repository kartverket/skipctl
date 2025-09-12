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

func (v *Validator) ValidateManifest(file *Document) error {
	switch file.Extension {
	case constants.ManifestSuffixJsonnet:
		if jerr := v.validateJsonnet(file); jerr != nil {
			return jerr
		}
		return nil
	case constants.ManifestSuffixYaml, constants.ManifestSuffixYml:
		if yerr := v.validateYaml(file); yerr != nil {
			return yerr
		}
		return nil
	}
	return nil
}
func (v *Validator) validateJsonnet(file *Document) error {
	// There is a memory corruption bug that leads to segfaults if we reuse the same VM for multiple evaluations.
	// if there is a syntax error within the Jsonnet file, the VM gets corrupted and cannot be used again.
	vm := jsonnet.MakeVM()

	node, err := jsonnet.SnippetToAST(file.Name, file.Content)
	if err != nil {
		v.res.ErrorCount++
		return err
	}

	content, err := vm.Evaluate(node)
	if err != nil {
		v.res.ErrorCount++
		return err
	}
	res, k8err := v.k8s.validateK8sSchema(file.Name, content)
	if k8err != nil {
		return k8err
	}
	v.countValidateRes(&res)
	return nil
}

func (v *Validator) validateYaml(d *Document) error {
	result, jerr := v.k8s.validateK8sSchema(d.Name, d.Content)
	if jerr != nil {
		return jerr
	}
	v.countValidateRes(&result)
	return nil
}
func (v *Validator) countValidateRes(result *ValidateResult) {
	v.res.ErrorCount += result.ErrorCount
	v.res.InvalidCount += result.InvalidCount
	v.res.SkippedCount += result.SkippedCount
	v.res.ValidCount += result.ValidCount
}

func (v *Validator) GetResults() *ValidateResult {
	return v.res
}
