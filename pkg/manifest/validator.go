package manifest

import (
	"log/slog"

	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/constants"
	"github.com/kartverket/skipctl/pkg/logging"
)

type Validator struct {
	k8s        *K8sValidator
	res        *ValidateResult
	outputJSON bool
	log        *slog.Logger
}

func NewValidator(tempDir string, outputJSON bool) *Validator {
	return &Validator{
		k8s:        NewK8sValidator(tempDir),
		res:        &ValidateResult{},
		outputJSON: outputJSON,
		log:        logging.RawLogger(),
	}
}

func (v *Validator) WithLogger(l *slog.Logger) *Validator {
	v.log = l
	return v
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
		logging.LogValidationErrorsTo(v.log, file.Name, nil, err, v.outputJSON)
		return err
	}

	content, err := vm.Evaluate(node)
	if err != nil {
		v.res.ErrorCount++
		logging.LogValidationErrorsTo(v.log, file.Name, nil, err, v.outputJSON)
		return err
	}
	res, results, k8err := v.k8s.validateK8sSchema(file.Name, content)

	// Log validation errors for each result
	for _, result := range results {
		logging.LogValidationErrorsTo(v.log, file.Name, result.ValidationErrors, result.Err, v.outputJSON)
	}

	v.countValidateRes(&res)
	return k8err
}

func (v *Validator) validateYaml(d *Document) error {
	result, results, jerr := v.k8s.validateK8sSchema(d.Name, d.Content)

	// Log validation errors for each result
	for _, r := range results {
		logging.LogValidationErrorsTo(v.log, d.Name, r.ValidationErrors, r.Err, v.outputJSON)
	}

	v.countValidateRes(&result)
	return jerr
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
