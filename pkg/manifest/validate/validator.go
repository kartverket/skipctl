package validate

import (
	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/manifest"
)

type Validator struct {
	k8s        *manifest.K8sValidator
	res        *manifest.ValidateResult
	outputJSON bool
}

func NewValidator(tempDir string, outputJSON bool) *Validator {
	return &Validator{
		k8s:        manifest.NewK8sValidator(tempDir),
		res:        &manifest.ValidateResult{},
		outputJSON: outputJSON,
	}
}

func (v *Validator) ValidateManifest(file *manifest.Document) error {
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
func (v *Validator) validateJsonnet(file *manifest.Document) error {
	// There is a memory corruption bug that leads to segfaults if we reuse the same VM for multiple evaluations.
	// if there is a syntax error within the Jsonnet file, the VM gets corrupted and cannot be used again.
	vm := jsonnet.MakeVM()

	node, err := jsonnet.SnippetToAST(file.Name, file.Content)
	if err != nil {
		v.res.ErrorCount++
		logging.LogValidationErrors(file.Name, nil, err, v.outputJSON)
		return err
	}

	content, err := vm.Evaluate(node)
	if err != nil {
		v.res.ErrorCount++
		logging.LogValidationErrors(file.Name, nil, err, v.outputJSON)
		return err
	}
	res, results, k8err := v.k8s.ValidateK8sSchema(file.Name, content)

	// Log validation errors for each result
	for _, result := range results {
		logging.LogValidationErrors(file.Name, result.ValidationErrors, result.Err, v.outputJSON)
	}

	v.countValidateRes(&res)
	return k8err
}

func (v *Validator) validateYaml(d *manifest.Document) error {
	result, results, jerr := v.k8s.ValidateK8sSchema(d.Name, d.Content)

	// Log validation errors for each result
	for _, r := range results {
		logging.LogValidationErrors(d.Name, r.ValidationErrors, r.Err, v.outputJSON)
	}

	v.countValidateRes(&result)
	return jerr
}
func (v *Validator) countValidateRes(result *manifest.ValidateResult) {
	v.res.ErrorCount += result.ErrorCount
	v.res.InvalidCount += result.InvalidCount
	v.res.SkippedCount += result.SkippedCount
	v.res.ValidCount += result.ValidCount
}

func (v *Validator) GetResults() *manifest.ValidateResult {
	return v.res
}
