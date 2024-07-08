package logging

import (
	"errors"
	"strings"
)

// OutputMode is a custom type for specifying a small number of allowed
// parameters in Cobra (cli-library)
type OutputMode string

const (
	OutputModeText    OutputMode = "text"
	OutputModeJSON    OutputMode = "json"
	OutputModeInvalid OutputMode = "invalid"
)

func ParseOutputMode(mode string) (OutputMode, error) {
	switch strings.ToLower(mode) {
	case "text":
		return OutputModeText, nil
	case "json":
		return OutputModeJSON, nil
	default:
		return OutputModeInvalid, errors.New("unknown output mode")
	}
}