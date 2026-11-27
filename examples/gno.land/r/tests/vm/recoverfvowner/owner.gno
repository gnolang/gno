// Package recoverfvowner holds objects whose methods panic, for the
// recover_fv file tests.
package recoverfvowner

import "gno.land/p/demo/tests/recoverfv"

type O struct{ n int }

func (o *O) Boom()  { panic("owner boom") }
func (o *O) Inc()   { o.n++ }
func (o *O) N() int { return o.n }

var (
	o = &O{}
	c = recoverfv.New()
)

func Get() *O { return o }

func GetC() *recoverfv.C { return c }
