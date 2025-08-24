package v2impl

var impl = &Impl{}

func Instance() *Impl {
	return impl
}

type Impl struct{}

func (i Impl) Render(path string) string { return "hello from v2" }
func (i Impl) Foo() error                { return nil }
