package pkg1

type IC interface {
	Inc()
}

type Counter struct {
	n int
}

func (c *Counter) Inc() {
	c.n++
}
