package manifest

// import (
// 	"errors"
// 	"log/slog"
// 	"os"

// 	"github.com/kartverket/skipctl/pkg/logging"
// )

// type Processor struct {
// 	log *slog.Logger
// }

// type StringToErrorFunc func(string) error

// type StringToValidateResultFunc func(string) (ValidateResult, error)

// func NewManifestProcessor() *Processor {
// 	return &Processor{
// 		log: logging.Logger(),
// 	}
// }

// func (p *Processor) ProcessManifests(files []string, process StringToErrorFunc) {
// 	failed := false

// 	for _, file := range files {
// 		err := process(file)

// 		if err != nil {
// 			p.log.Error("failed", "file", file, "error", err.Error())
// 			failed = true
// 		}
// 	}
// 	if failed {
// 		os.Exit(1)
// 	}
// }

// func (p *Processor) ProcessValidationManifests(files []string, process StringToValidateResultFunc) error {
// 	failed := false
// 	var validCount, invalidCount, errorCount, skippedCount int

// 	for _, file := range files {
// 		summary, validationErr := process(file)
// 		if validationErr != nil {
// 			p.log.Error("validation failed", "file", file, "error", validationErr.Error())
// 			failed = true
// 		}

// 		validCount += summary.ValidCount
// 		invalidCount += summary.InvalidCount
// 		errorCount += summary.ErrorCount
// 		skippedCount += summary.SkippedCount
// 	}

// 	totalResources := validCount + invalidCount + errorCount + skippedCount

// 	p.log.Info("validation completed", "totalResources", totalResources, "valid", validCount, "invalid", invalidCount, "errors", errorCount, "skipped", skippedCount)

// 	if failed {
// 		return errors.New("validation failed")
// 	}

// 	return nil
// }
