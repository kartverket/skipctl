package manifest

import (
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

type ManifestProcessor struct {
	log *slog.Logger
}

type ManifestFileToErrorFunc func(*utils.ManifestFile) error

func NewManifestFileProcessor() *ManifestProcessor {
	return &ManifestProcessor{
		log: logging.Logger(),
	}
}

func (p *ManifestProcessor) ProcessManifestFiles(files []*utils.ManifestFile, process ManifestFileToErrorFunc) {
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
