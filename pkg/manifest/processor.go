package manifest

import (
	"log/slog"
	"os"

	"github.com/kartverket/skipctl/pkg/logging"
)

type ManifestProcessor struct {
	log      *slog.Logger
	logLabel string
}

type StringToErrorFunc func(string) error

func NewManfiestProcessor(label string) *ManifestProcessor {
	return &ManifestProcessor{
		log:      logging.Logger(),
		logLabel: label,
	}
}

func (p *ManifestProcessor) ProcessManifests(files []string, process StringToErrorFunc) {
	failed := false

	for _, file := range files {
		err := process(file)

		if err != nil {
			p.log.Error(p.logLabel+" failed",
				"file", file,
				"error", err.Error(),
			)
			failed = true
		} else {
			p.log.Info(p.logLabel+" succeeded",
				"file", file,
			)
		}
	}
	if failed {
		os.Exit(1)
	}
}
