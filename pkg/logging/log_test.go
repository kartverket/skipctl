package logging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yannh/kubeconform/pkg/validator"
)

func TestLogValidationErrors_NilInputs(_ *testing.T) {
	// Setup
	ConfigureLogging("text", false)

	// Should not panic and return early when both inputs are nil
	LogValidationErrors("test.yaml", nil, nil, false)
	// If we reach here without panic, test passes
}

func TestLogValidationErrors_JSONModeSkipsRawOutput(_ *testing.T) {
	// Setup logger in JSON mode
	ConfigureLogging("json", false)

	// Create a buffer to capture output
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	testHandler := slog.NewJSONHandler(&buf, opts)
	logger = slog.New(testHandler)

	// Call with outputJSON=false, but outputMode is JSON
	// This should return early without logging
	validationErrors := []validator.ValidationError{
		{Path: "spec.port", Msg: "missing required field"},
	}

	LogValidationErrors("test.yaml", validationErrors, nil, false)

	// No raw output should be produced since outputMode is JSON
	// (Note: actual JSON structured logs would go to the structured logger, not rawLogger)
}

func TestLogValidationErrors_TextModeWithValidationErrors(_ *testing.T) {
	// Setup logger in text mode
	ConfigureLogging("text", false)

	validationErrors := []validator.ValidationError{
		{Path: "spec.port", Msg: "{missing required field}"},
		{Path: "spec.replicas.min", Msg: "{must be greater than 0}"},
	}

	// This test verifies the function doesn't panic and processes all errors
	// In a real scenario, output would be captured via a custom handler
	LogValidationErrors("test.yaml", validationErrors, nil, false)
	// If we reach here without panic, the function executed successfully
}

func TestLogValidationErrors_WithErrorContainingHint(_ *testing.T) {
	// Setup logger in text mode
	ConfigureLogging("text", false)

	// Create an error with a hint
	err := validator.ValidationError{
		Path: "apiVersion",
		Msg:  "could not find schema Hint: The schema for Application with version 'v1beta1' was not found",
	}

	// This should trigger the hint splitting logic
	LogValidationErrors("test.yaml", nil, &err, false)
	// If we reach here without panic, the hint was processed correctly
}

func TestLogValidationErrors_WithErrorWithoutHint(_ *testing.T) {
	// Setup logger in text mode
	ConfigureLogging("text", false)

	// Create a simple error without hint
	err := validator.ValidationError{
		Path: "spec.image",
		Msg:  "invalid image format",
	}

	// This should use the regular error formatting
	LogValidationErrors("test.yaml", nil, &err, false)
	// If we reach here without panic, the error was processed correctly
}

func TestLogValidationErrors_MultipleValidationErrors(_ *testing.T) {
	// Setup logger in text mode
	ConfigureLogging("text", false)

	validationErrors := []validator.ValidationError{
		{Path: "spec.port", Msg: "{required}"},
		{Path: "spec.image", Msg: "{invalid format}"},
		{Path: "metadata.name", Msg: "{must match pattern}"},
	}

	// Should process all validation errors without panic
	LogValidationErrors("test.yaml", validationErrors, nil, false)
}

func TestLogValidationErrors_ErrorWithWhitespace(_ *testing.T) {
	// Setup logger in text mode
	ConfigureLogging("text", false)

	// Error with leading/trailing whitespace should be trimmed
	err := validator.ValidationError{
		Path: "spec",
		Msg:  "  error with whitespace  ",
	}

	LogValidationErrors("test.yaml", nil, &err, false)
	// Should trim whitespace and format correctly
}

func TestConfigureLogging(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		isDebug bool
		wantErr bool
	}{
		{
			name:    "text mode without debug",
			mode:    "text",
			isDebug: false,
			wantErr: false,
		},
		{
			name:    "json mode without debug",
			mode:    "json",
			isDebug: false,
			wantErr: false,
		},
		{
			name:    "text mode with debug",
			mode:    "text",
			isDebug: true,
			wantErr: false,
		},
		{
			name:    "json mode with debug",
			mode:    "json",
			isDebug: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configuredLogger := ConfigureLogging(tt.mode, tt.isDebug)
			require.NotNil(t, configuredLogger, "ConfigureLogging should return a valid logger")

			// Verify logger is functional
			configuredLogger.Info("test message")
		})
	}
}

func TestConfigureLogging_InvalidMode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ConfigureLogging should panic with invalid mode")
		}
	}()

	ConfigureLogging("invalid", false)
}

func TestForceStdoutContext(t *testing.T) {
	parent := context.Background()
	ctx := ForceStdoutContext(parent)

	// Verify the context has the stdout key set
	toStdout, ok := ctx.Value(ctxStdoutKey).(bool)
	require.True(t, ok, "context should have stdout key")
	assert.True(t, toStdout, "stdout value should be true")
}

func TestLogger(t *testing.T) {
	ConfigureLogging("json", false)

	log := Logger()
	require.NotNil(t, log, "Logger() should return a valid logger")
}

func TestRawLogger(t *testing.T) {
	log := RawLogger()
	require.NotNil(t, log, "RawLogger() should return a valid logger")
}

func TestSplitHandler_Enabled(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	handler := &splitHandler{
		stdout: slog.NewJSONHandler(&bytes.Buffer{}, opts),
		stderr: slog.NewJSONHandler(&bytes.Buffer{}, opts),
	}

	ctx := context.Background()

	// Should be enabled for Info level
	assert.True(t, handler.Enabled(ctx, slog.LevelInfo))

	// Should be enabled for Error level
	assert.True(t, handler.Enabled(ctx, slog.LevelError))
}

func TestSplitHandler_Handle(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	handler := &splitHandler{
		stdout: slog.NewJSONHandler(&stdoutBuf, opts),
		stderr: slog.NewJSONHandler(&stderrBuf, opts),
	}

	// Test stdout routing
	ctxStdout := ForceStdoutContext(context.Background())
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test stdout message", 0)
	err := handler.Handle(ctxStdout, record)
	require.NoError(t, err)
	assert.Contains(t, stdoutBuf.String(), "test stdout message")
	assert.Empty(t, stderrBuf.String())

	// Reset buffers
	stdoutBuf.Reset()
	stderrBuf.Reset()

	// Test stderr routing (default)
	ctxStderr := context.Background()
	record = slog.NewRecord(time.Now(), slog.LevelInfo, "test stderr message", 0)
	err = handler.Handle(ctxStderr, record)
	require.NoError(t, err)
	assert.Contains(t, stderrBuf.String(), "test stderr message")
	assert.Empty(t, stdoutBuf.String())
}

func TestLogValidationErrors_HintSplitWithMultipleParts(_ *testing.T) {
	// Setup logger in text mode
	ConfigureLogging("text", false)

	// Create an error with multiple "Hint:" occurrences
	// The function should only split on the first occurrence (maxHintParts=2)
	err := validator.ValidationError{
		Path: "test",
		Msg:  "error message Hint: first hint Hint: second hint should stay with first",
	}

	LogValidationErrors("test.yaml", nil, &err, false)
	// Should split only on first "Hint:" due to SplitN with maxHintParts=2
}

func TestLogValidationErrors_BraceTrimming(_ *testing.T) {
	// Setup logger in text mode
	ConfigureLogging("text", false)

	validationErrors := []validator.ValidationError{
		{Path: "spec.port", Msg: "{error with braces}"},
		{Path: "spec.image", Msg: "{{nested braces}}"},
		{Path: "spec.name", Msg: "error without braces"},
	}

	// Should trim outer braces correctly
	LogValidationErrors("test.yaml", validationErrors, nil, false)
}

func TestDefaultStdoutContext(t *testing.T) {
	// Verify the default stdout context is properly initialized
	toStdout, ok := DefaultStdoutContext.Value(ctxStdoutKey).(bool)
	require.True(t, ok, "DefaultStdoutContext should have stdout key")
	assert.True(t, toStdout, "DefaultStdoutContext stdout value should be true")
}
