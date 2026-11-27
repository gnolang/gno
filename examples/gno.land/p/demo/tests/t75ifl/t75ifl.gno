package t75ifl

type Eacher interface{ Each(xs []int, f func(int)) }

type Impl struct{}

func (Impl) Each(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}
