package manifest

import (
	"log/slog"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/kartverket/skipctl/pkg/logging"
)

type JsonnetValidator struct {
	vm  *jsonnet.VM
	log *slog.Logger
}

func NewJsonnetValidator() *JsonnetValidator {

	return &JsonnetValidator{
		vm:  jsonnet.MakeVM(),
		log: logging.Logger(),
	}
}

func (v *JsonnetValidator) ValidateManifest(filepath string) error {

	_, err := v.vm.EvaluateFile(filepath)

	if err != nil {
		// TODO Find a better solution for this
		if strings.Contains(err.Error(), "Top-level function call") {
			return nil
		}
		return err
	}
	return nil
}
