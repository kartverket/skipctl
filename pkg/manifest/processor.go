package manifest

import (
	"errors"
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/logging"
)

type Processor struct {
	log *slog.Logger
}
type ProcessorFunc struct {
}
type FileToErrorFunc func(*Document) error

func NewDocumentProcessor() *Processor {
	return &Processor{
		log: logging.Logger(),
	}
}

func (p *Processor) ProcessDocuments(files []*Document, process FileToErrorFunc) {
	failed := false

	for _, file := range files {
		err := process(file)

		if err != nil {
			p.log.Error("failed", "file", file.Name, "error", err.Error())
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

func (p *Processor) ProcessValidationManifests(files []*Document, process FileToErrorFunc) error {
	for _, file := range files {
		validationErr := process(file)
		if validationErr != nil {
			p.log.Error("validation failed", "file", file, "error", validationErr.Error())
		}
	}
	totalResources := validateResult.ErrorCount + validateResult.ValidCount + validateResult.InvalidCount + validateResult.SkippedCount
	p.log.Info("validation completed", "totalResources", totalResources, "valid", validateResult.ValidCount, "invalid", validateResult.InvalidCount, "errors", validateResult.ErrorCount, "skipped", validateResult.SkippedCount)

	if validateResult.ErrorCount > 0 || validateResult.InvalidCount > 0 {
		return errors.New("validation failed")
	}

	return nil
}
