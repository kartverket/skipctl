package manifest

import (
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/logging"
)

type Processor struct {
	log *slog.Logger
}

type StringToErrorFunc func(string) error

func NewManifestProcessor() *Processor {
	return &Processor{
		log: logging.Logger(),
	}
}

func (p *Processor) ProcessManifests(files []string, process StringToErrorFunc) {
	failed := false

	for _, file := range files {
		err := process(file)

		if err != nil {
			p.log.Error("failed", "file", file, "error", err.Error())
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}
