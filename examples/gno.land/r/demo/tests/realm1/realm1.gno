package realm1

import "gno.land/p/demo/tests/pkg1"

var c = Counter{}

var c1 = pkg1.Counter{}

type Counter struct {
	n int
}

func (c *Counter) Inc() {
	c.n++
}

func GetCounter() *Counter {
	return &c
}

func GetCounterPkg() *pkg1.Counter {
	return &c1
}
