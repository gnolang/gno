package v1impl

import "errors"

var impl = &Impl{}

func Instance() *Impl {
	return impl
}

type Impl struct{}

func (i Impl) Render(path string) string { return "hello from v1" }
func (i Impl) Foo() error                { return errors.New("not implemented") }
