package ifacewrap

import "gno.land/p/demo/tests/launderpkg/v0"

type Wrap struct{ T *launderpkg.Object }

func New(o *launderpkg.Object) *Wrap { return &Wrap{o} }

func (w *Wrap) Put(s string) { w.T.Set(s) }

type WrapV struct{ T *launderpkg.Object }

func (w WrapV) Put(s string) { w.T.Field = s }
