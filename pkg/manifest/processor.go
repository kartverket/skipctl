package manifest

import (
	"errors"
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

	for _, file := range files {
		err := process(file)

		if err != nil {
			p.log.Error("failed", "file", file.Name, "error", err.Error())
			failed = true
		}
	}
	if failed {
		return errors.New("error while processing documents")
	}
	return nil
}
