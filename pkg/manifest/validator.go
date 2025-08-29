package manifest

import (
	"github.com/google/go-jsonnet"
)

type JsonnetValidator struct {
	vm *jsonnet.VM
}

func NewJsonnetValidator() *JsonnetValidator {
	return &JsonnetValidator{
		vm: jsonnet.MakeVM(),
	}
}

func (v *JsonnetValidator) ValidateManifest(filepath string) error {
	_, err := v.vm.EvaluateFile(filepath)
	return err
}
