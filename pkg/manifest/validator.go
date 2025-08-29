package manifest

import (
	"log/slog"

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

	return err
}
