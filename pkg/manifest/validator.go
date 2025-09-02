package manifest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"

	"github.com/kartverket/skipctl/pkg/constants"
)

type Validator struct {
	jsonnet *jsonnet.VM
}

func NewValidator() *Validator {
	return &Validator{
		jsonnet: jsonnet.MakeVM(),
	}
}

func (v *Validator) ValidateManifest(filename string) error {
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension { //nolint:gocritic // singleCaseSwitch: this is intentional
	case constants.ManifestSuffixJsonnet:
		return v.validateJsonnet(filename)
	case constants.ManifestSuffixLibsonnet:
		return v.validateJsonnet(filename)
	}
	return nil
}

func (v *Validator) validateJsonnet(filename string) error {
	_, err := v.jsonnet.EvaluateFile(filename)
	return err
}

func validateJsonnetSyntax(filename string) error {
	// Read the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Strip trailing newlines (\n and \r)
	trimmed := strings.TrimRight(string(content), "\r\n")

	// Parse the content using SnippetToAST
	_, err = jsonnet.SnippetToAST(filename, trimmed)
	return err // Returns nil if syntax is valid, error if not

}
