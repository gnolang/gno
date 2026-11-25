// Package fvshare is a test fixture for R2a: each entry makes its own writer
// (vault.Bump) and then calls a function value a caller's handler chose.
package fvshare

type Vault struct{ N int }

var vault = &Vault{}

func (v *Vault) Bump() { v.N++ }

func Count() int { return vault.N }

type Hooks struct{ OnDone func() }

var hooks = &Hooks{}

func HooksPtr() *Hooks { return hooks }

// Private makes a writer and uses it itself: it writes.
func Private(cur realm) {
	f := vault.Bump
	fs := []func(){f}
	fs[0]()
}

var wrap = func(f func()) func() { return f }

// Decorate passes a fresh writer through its own stored helper: a function
// called through a value is read-only code, so the writer comes back read-only.
func Decorate(cur realm) {
	cb := wrap(vault.Bump)
	cb()
}

type PtrPicker interface{ PickPtr() *func() }

// SettlePtr stores its hook, then calls through a pointer the handler picked.
func SettlePtr(cur realm, p PtrPicker) {
	hooks.OnDone = vault.Bump
	q := p.PickPtr()
	(*q)()
}

// SettleArg stores its hook, then calls through a pointer argument.
func SettleArg(cur realm, q *func()) {
	hooks.OnDone = vault.Bump
	(*q)()
}

type Echoer interface{ Echo(p *func()) *func() }

// SettleLocal shows a local slot, fills it, then calls through the pointer
// the handler returned.
func SettleLocal(cur realm, e Echoer) {
	var f func()
	q := e.Echo(&f)
	f = vault.Bump
	(*q)()
}

type Ctx struct {
	Undo func()
	Cb   func()
}

type CtxPicker interface{ Pick(c *Ctx) func() }

// ProcessCtx shows a context holding its undo and runs the handler's choice.
func ProcessCtx(cur realm, h CtxPicker) {
	cb := h.Pick(&Ctx{Undo: vault.Bump})
	cb()
}

type Callbacker interface{ Callback(undo func()) func() }

// ProcessUndo hands its undo by value and runs what the handler returns.
func ProcessUndo(cur realm, h Callbacker) {
	cb := h.Callback(vault.Bump)
	cb()
}

// ProcessUndoFn is ProcessUndo with the handler as a function value.
func ProcessUndoFn(cur realm, h func(undo func()) func()) {
	cb := h(vault.Bump)
	cb()
}

type Getter interface{ Use(get func() func()) func() }

// ProcessCapture shows a getter closure that captures its undo.
func ProcessCapture(cur realm, h Getter) {
	w := vault.Bump
	cb := h.Use(func() func() { return w })
	cb()
}

type holder struct{ f func() }

func (hd *holder) F() func() { return hd.f }

// ProcessMethodValue shows a method value whose receiver holds its undo.
func ProcessMethodValue(cur realm, h Getter) {
	hd := &holder{f: vault.Bump}
	cb := h.Use(hd.F)
	cb()
}

type Filler interface{ Fill(c *Ctx) }

// ProcessDefer lets a deferred handler fill in a callback, then runs it.
func ProcessDefer(cur realm, h Filler) {
	c := &Ctx{Undo: vault.Bump}
	defer func() {
		if c.Cb != nil {
			c.Cb()
		}
	}()
	defer h.Fill(c)
}

type VPicker interface{ PickV(fs ...func()) func() }

// ProcessVariadic offers options and runs the handler's choice.
func ProcessVariadic(cur realm, h VPicker) {
	cb := h.PickV(func() {}, vault.Bump)
	cb()
}

// ProcessSpread is ProcessVariadic with a ready slice.
func ProcessSpread(cur realm, h VPicker) {
	opts := []func(){func() {}, vault.Bump}
	cb := h.PickV(opts...)
	cb()
}

// Take accepts any value and does nothing with it.
func Take(cur realm, v any) {}
