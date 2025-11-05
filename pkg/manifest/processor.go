package manifest

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/logging"
)

type Processor struct {
	log *slog.Logger
}

type FileToErrorFunc func(*Document) error

func NewDocumentProcessor() *Processor {
	return &Processor{
		log: logging.Logger(),
	}
}

func (p *Processor) ProcessDocuments(files []*Document, process FileToErrorFunc) error {
	failed := false
	errs := []error{}

	for _, file := range files {
		err := process(file)

		if err != nil {
			errs = append(errs, fmt.Errorf("\033[37;41mError\033[0m" + " while processing document %s: %w ", file.Name, err + "\n"))
			failed = true
		}
	}
	if failed {
		return errors.Join(errs...)
	}
	return nil
}
