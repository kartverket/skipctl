package manifest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yannh/kubeconform/pkg/validator"
)

func TestCheckIfValidSchema_NilError(t *testing.T) {
	k8 := &K8sValidator{}
	result := validator.Result{
		Err: nil,
	}

	err := k8.checkIfValidSchema(result)
	assert.NoError(t, err, "should return nil when input error is nil")
}

func TestCheckIfValidSchema_NonSchemaError(t *testing.T) {
	k8 := &K8sValidator{}
	originalErr := errors.New("some other validation error")
	result := validator.Result{
		Err: originalErr,
	}

	err := k8.checkIfValidSchema(result)
	assert.Equal(t, originalErr, err, "should return original error when not a schema error")
}

func TestCheckIfValidSchema_SchemaErrorWithoutResource(t *testing.T) {
	k8 := &K8sValidator{}
	originalErr := errors.New("could not find schema for Application")
	result := validator.Result{
		Err: originalErr,
		// Resource is not set, so signature extraction should fail
	}

	err := k8.checkIfValidSchema(result)
	// Should return original error since we can't extract signature
	assert.Equal(t, originalErr, err, "should return original error when signature extraction fails")
}

func TestIsJSONArray(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "valid JSON array",
			content:  `[{"key": "value"}]`,
			expected: true,
		},
		{
			name:     "empty JSON array",
			content:  `[]`,
			expected: true,
		},
		{
			name:     "JSON object",
			content:  `{"key": "value"}`,
			expected: false,
		},
		{
			name:     "string starting with bracket",
			content:  `[not an array`,
			expected: false,
		},
		{
			name:     "string ending with bracket",
			content:  `not an array]`,
			expected: false,
		},
		{
			name:     "empty string",
			content:  ``,
			expected: false,
		},
		{
			name:     "whitespace with array",
			content:  ` [1, 2, 3] `,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJSONArray(tt.content)
			assert.Equal(t, tt.expected, result, "unexpected result for %s", tt.name)
		})
	}
}

func TestValidateResult_GetTotalResources(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidateResult
		expected int
	}{
		{
			name: "all zeros",
			result: ValidateResult{
				ValidCount:   0,
				InvalidCount: 0,
				ErrorCount:   0,
				SkippedCount: 0,
			},
			expected: 0,
		},
		{
			name: "mixed counts",
			result: ValidateResult{
				ValidCount:   5,
				InvalidCount: 2,
				ErrorCount:   1,
				SkippedCount: 3,
			},
			expected: 11,
		},
		{
			name: "only valid",
			result: ValidateResult{
				ValidCount: 10,
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.result.GetTotalResources()
			assert.Equal(t, tt.expected, total, "unexpected total resources")
		})
	}
}

func TestValidateResult_HasValidationFailed(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidateResult
		expected bool
	}{
		{
			name: "no failures",
			result: ValidateResult{
				ValidCount:   5,
				SkippedCount: 2,
			},
			expected: false,
		},
		{
			name: "has errors",
			result: ValidateResult{
				ValidCount: 5,
				ErrorCount: 1,
			},
			expected: true,
		},
		{
			name: "has invalid",
			result: ValidateResult{
				ValidCount:   5,
				InvalidCount: 1,
			},
			expected: true,
		},
		{
			name: "has both errors and invalid",
			result: ValidateResult{
				ErrorCount:   2,
				InvalidCount: 3,
			},
			expected: true,
		},
		{
			name:     "all zeros",
			result:   ValidateResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failed := tt.result.HasValidationFailed()
			assert.Equal(t, tt.expected, failed, "unexpected validation failure status")
		})
	}
}
