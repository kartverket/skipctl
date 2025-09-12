package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/crd"
	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
	"github.com/yannh/kubeconform/pkg/validator"
)

type K8sValidator struct {
	log       *slog.Logger
	validator validator.Validator
}

type ValidateResult struct {
	ValidCount   int
	InvalidCount int
	ErrorCount   int
	SkippedCount int
}

func (vr *ValidateResult) GetTotalResources() int {
	return vr.ErrorCount + vr.ValidCount + vr.InvalidCount + vr.SkippedCount
}

func (vr *ValidateResult) HasValidationFailed() bool {
	return vr.ErrorCount > 0 || vr.InvalidCount > 0
}

// isJSONArray checks if the content is a JSON array.
func isJSONArray(content string) bool {
	return strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]")
}

// validateK8sSchema validates the Kubernetes schema of the given content.
//
// The content can be either in form JSON or YAML.
// It handles both single resources and arrays of resources.
func (k8 *K8sValidator) validateK8sSchema(filename string, content string) (ValidateResult, error) {
	content = strings.TrimSpace(content)

	if isJSONArray(content) {
		// Parse the content as an array of raw JSON messages
		var resources []json.RawMessage
		if err := json.Unmarshal([]byte(content), &resources); err != nil {
			return ValidateResult{}, fmt.Errorf("failed to parse JSON array: %w", err)
		}

		// Validate each resource in the array
		return k8.validateResourceArray(filename, resources)
	}

	reader := io.NopCloser(strings.NewReader(content))
	results := k8.validator.Validate(filename, reader)
	return k8.processValidationResults(filename, results)
}

// validateResourceArray validates each resource in a JSON array.
//
// For each resource, it creates a reader and validates it individually.
// Only for JSON arrays.
func (k8 *K8sValidator) validateResourceArray(filename string, resources []json.RawMessage) (ValidateResult, error) {
	var allResults []validator.Result

	for i, resource := range resources {
		reader := io.NopCloser(strings.NewReader(string(resource)))
		docFilename := fmt.Sprintf("%s[%d]", filename, i)
		results := k8.validator.Validate(docFilename, reader)
		allResults = append(allResults, results...)
	}

	return k8.processValidationResults(filename, allResults)
}

// processValidationResults processes and logs the results of the validation.
//
// Counts the number of valid, invalid, error, and skipped resources,
// and returns the errors encountered during validation.
func (k8 *K8sValidator) processValidationResults(filename string, results []validator.Result) (ValidateResult, error) {
	// Initialize counters for each status
	var validCount, invalidCount, errorCount, skippedCount int

	for _, result := range results {
		switch result.Status {
		case validator.Valid:
			validCount++

		case validator.Invalid:
			invalidCount++

			k8.log.Error("file is invalid", "filename", filename, "errors", result.ValidationErrors)

		case validator.Error:
			errorCount++

			k8.log.Error("error processing resource", "filename", filename, "error", result.Err.Error())

		case validator.Skipped:
			skippedCount++

		case validator.Empty:
			// Skip empty documents
		}
	}

	return ValidateResult{
		ValidCount:   validCount,
		InvalidCount: invalidCount,
		ErrorCount:   errorCount,
		SkippedCount: skippedCount,
	}, nil
}

// initValidator initializes the Kubernetes schema validator.
//
// Enforces strict validation mode and loads the necessary custom K8s CRDs.
func initValidator(tempDir string) validator.Validator {
	log := logging.Logger()

	err := utils.CopyFilesToDirectory(crd.Schemas, tempDir)
	if err != nil {
		log.Error("Failed to copy embedded schema files", "error", err)
		os.Exit(1)
	}

	crdPath := tempDir + "/schemas" + "/{{ .Group }}_{{ .ResourceAPIVersion }}_{{ .ResourceKind }}.json"
	schemaLocations := []string{
		"https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/{{ .NormalizedKubernetesVersion }}-standalone{{ .StrictSuffix }}/{{ .ResourceKind }}{{ .KindSuffix }}.json",
		crdPath,
	}

	v, err := validator.New(schemaLocations, validator.Opts{Strict: true})
	if err != nil {
		log.Error("Failed to initialize K8s validator", "error", err)
		return nil
	}

	return v
}

func NewK8sValidator(tempDir string) *K8sValidator {
	return &K8sValidator{
		log:       logging.Logger(),
		validator: initValidator(tempDir),
	}
}
