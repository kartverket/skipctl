package manifest

import (
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

type Processor struct {
	log *slog.Logger
}

type FileToErrorFunc func(*utils.ManifestFile) error

func NewManifestFileProcessor() *Processor {
	return &Processor{
		log: logging.Logger(),
	}
}

func (p *Processor) ProcessManifestFiles(files []*utils.ManifestFile, process FileToErrorFunc) {
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
