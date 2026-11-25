// Package fvpass is a test fixture: library handlers that hand a caller's
// function value straight back.
package fvpass

type L struct{}

func (L) Callback(undo func()) func()  { return undo }
func (L) Use(get func() func()) func() { return get() }
func (L) PickV(fs ...func()) func()    { return fs[len(fs)-1] }

func Ident(f func()) func() { return f }

// Unwrapper is a foreign struct that shows the writer it holds.
type Unwrapper interface{ Get() func() }

// Catch recovers a thrown value and calls it once, keeping nothing.
func (L) Catch() {
	r := recover()
	switch v := r.(type) {
	case func():
		v()
	case []func():
		v[0]()
	case Unwrapper:
		v.Get()()
	}
}

// Rethrow is a catcher that recovers a thrown value and throws it on.
type Rethrow struct{}

func (Rethrow) Catch() { panic(recover()) }

// Keeper keeps a recovered value behind a getter closure, whose captures are
// never copied, and hands it back from Pick.
type Keeper struct {
	put func(g func() any)
	get func() any
}

func (L) NewKeeper() any {
	var stash func() any
	return &Keeper{
		put: func(g func() any) { stash = g },
		get: func() any {
			if stash == nil {
				return nil
			}
			return stash()
		},
	}
}

func (k *Keeper) Catch() {
	v := recover()
	k.put(func() any { return v })
}

func (k *Keeper) Pick() func() {
	f, _ := k.get().(func())
	return f
}
