package httpx

import "github.com/go-playground/validator/v10"

type Validator interface {
	Validate(any) error
	Underlying() any
}

var defaultValidator = &validate{inner: validator.New()}

type validate struct {
	inner *validator.Validate
}

func (v *validate) Validate(val any) error {
	return v.inner.Struct(val)
}

func (v *validate) Underlying() any {
	return v.inner
}
