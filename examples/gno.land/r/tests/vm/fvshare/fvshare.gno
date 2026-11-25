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

var (
	onSale  func()
	saleMap = map[string]func(){}
	saleCap = make([]func(), 0, 4)
	saleNew []func()
	saleDst = make([]func(), 1)
)

// StoreLocal stores a writer by the named route, then runs its local name.
func StoreLocal(cur realm, route string) {
	f := vault.Bump
	switch route {
	case "var":
		onSale = f
	case "map":
		saleMap["a"] = f
	case "append in place":
		saleCap = append(saleCap, f)
	case "append grown":
		saleNew = append(saleNew, f)
	case "copy":
		copy(saleDst, []func(){f})
	}
	f()
}

// StoreBox stores a writer, then runs the stored one.
func StoreBox(cur realm) {
	f := vault.Bump
	onSale = f
	onSale()
}

// StoreAlias stores a writer, then runs a local read back from the box.
func StoreAlias(cur realm) {
	onSale = vault.Bump
	g := onSale
	g()
}

// HandLocal hands a writer by value, then runs its local name.
func HandLocal(cur realm, h Callbacker) {
	f := vault.Bump
	h.Callback(f)
	f()
}

// HandStruct shows a struct holding the writer, then runs the field or the local.
func HandStruct(cur realm, h CtxPicker, field bool) {
	f := vault.Bump
	c := &Ctx{Undo: f}
	h.Pick(c)
	if field {
		c.Undo()
	} else {
		f()
	}
}

// ShownVar shows the variable itself, then runs it.
func ShownVar(cur realm, e Echoer) {
	f := vault.Bump
	e.Echo(&f)
	f()
}

func shownResult(e Echoer, out **func()) (f func()) {
	*out = e.Echo(&f)
	return vault.Bump
}

func shownResultDeferred(e Echoer, out **func()) (f func()) {
	defer func() {}()
	*out = e.Echo(&f)
	return vault.Bump
}

// ShownResult shows a named result before the return fills it.
func ShownResult(cur realm, e Echoer, deferred bool) {
	var q *func()
	if deferred {
		shownResultDeferred(e, &q)
	} else {
		shownResult(e, &q)
	}
	(*q)()
}

type KeyPicker interface {
	PickKeys(m map[*Ctx]bool) map[*Ctx]bool
}

// PickKeys runs the undo of each key the handler picked from a set.
func PickKeys(cur realm, h KeyPicker) {
	m := map[*Ctx]bool{&Ctx{Undo: func() {}}: true, &Ctx{Undo: vault.Bump}: true}
	for k := range h.PickKeys(m) {
		k.Undo()
	}
}

// HandKeys shows a set keyed by its context, then runs the context's undo.
func HandKeys(cur realm, h KeyPicker) {
	c := &Ctx{Undo: vault.Bump}
	h.PickKeys(map[*Ctx]bool{c: true})
	c.Undo()
}

type Planter interface {
	Plant(getBox func() *Ctx, getW func() func())
}

// ProcessPlant offers getters for a private box and a writer, then runs the box.
func ProcessPlant(cur realm, h Planter) {
	w := vault.Bump
	box := &Ctx{}
	h.Plant(func() *Ctx { return box }, func() func() { return w })
	if box.Cb != nil {
		box.Cb()
	}
}

var (
	keep func() func()
	kept func()
)

// MakeKeep stores a getter whose capture holds a fresh writer.
func MakeKeep(cur realm) {
	w := vault.Bump
	keep = func() func() { return w }
}

// UseKeep, a later crossing call, stores the writer the getter returns.
func UseKeep(cur realm) { kept = keep() }

// ClearKeep drops the getter; kept still holds the writer.
func ClearKeep(cur realm) { keep = nil }

type Runner interface{ Run(f func()) }

// RecoverRun makes a writer, lets the stranger's library run a closure that
// throws it from read-only code, and recovers it in writable victim code.
func RecoverRun(cur realm, h Runner) func() {
	w := vault.Bump
	var got func()
	func() {
		defer func() {
			if v := recover(); v != nil {
				got = v.(func())
			}
		}()
		h.Run(func() { panic(w) })
	}()
	got()
	return got
}

var nestedSaved func()

// SaveClose stores a writer in saved state.
func SaveClose(cur realm) { nestedSaved = vault.Bump }

// NestedSave saves a writer, crosses into the stranger, then calls the saved
// copy in the same crossing call.
func NestedSave(cur realm, cb func(cur realm)) {
	w := vault.Bump
	nestedSaved = w
	cb(cross(cur))
	nestedSaved()
}

// UseSaved runs the writer saved before the nested call.
func UseSaved(cur realm) { nestedSaved() }

// NestedCallBack crosses into the stranger, which crosses back here.
func NestedCallBack(cur realm, cb func(cur realm)) { cb(cross(cur)) }

// Call runs a function value a stranger kept.
func Call(cur realm, f func()) { f() }

// CallAny runs a function value a stranger kept in any shape.
func CallAny(cur realm, v any) { callAny(v) }

type DeferUser interface {
	Use(v any, set func(any))
}

type F func()

func (f F) Call() { f() }

func callAny(v any) {
	switch x := v.(type) {
	case func():
		x()
	case []func():
		x[0]()
	case *Ctx:
		x.Undo()
	case F:
		x()
	}
}

// DeferArgs defers one call that hands out a writer in the named shape, then
// calls what the stranger handed back.
func DeferArgs(cur realm, h DeferUser, other func(cur realm, v any), shape string) {
	w := vault.Bump
	var got any
	set := func(v any) { got = v }
	func() {
		switch shape {
		case "direct":
			defer h.Use(w, set)
		case "struct":
			defer h.Use(&Ctx{Undo: w}, set)
		case "slice":
			defer h.Use([]func(){w}, set)
		case "method":
			defer h.Use(F(w).Call, set)
		case "crossing":
			defer func() { other(cross(cur), w) }()
		}
	}()
	callAny(got)
}

type Catcher interface{ Catch() }

// WBox is a fresh struct holding a writer thrown from writable code.
type WBox struct{ F func() }

func (b *WBox) Get() func() { return b.F }

// ThrowLib panics with a fresh writer from writable code, while the stranger's
// deferred catcher calls it at once and keeps nothing.
func ThrowLib(cur realm, h Catcher, shape string) {
	w := vault.Bump
	var v any
	switch shape {
	case "direct":
		v = w
	case "struct":
		v = &WBox{F: w}
	case "slice":
		v = []func(){w}
	}
	func() {
		defer h.Catch()
		panic(v)
	}()
}

// Throw panics with a fresh writer from writable code, recovering nothing.
func Throw(cur realm) {
	w := vault.Bump
	panic(w)
}

type Keeper interface {
	Catch()
	Pick() func()
}

// ChalPick throws a fresh writer from writable code into the stranger's
// deferred catcher, then calls what the stranger picks.
func ChalPick(cur realm, h Keeper) {
	func() {
		defer h.Catch()
		w := vault.Bump
		panic(w)
	}()
	h.Pick()()
}

// ChalSlot throws a fresh writer from writable code into the stranger's
// deferred catcher, which may write it into a box the victim showed earlier.
func ChalSlot(cur realm, h interface {
	Show(c *Ctx)
	Catch()
}) {
	c := &Ctx{}
	h.Show(c)
	func() {
		defer h.Catch()
		w := vault.Bump
		panic(w)
	}()
	if c.Cb != nil {
		c.Cb()
	}
}

// RecoverKeep is RecoverRun that returns what it recovered without calling it.
func RecoverKeep(cur realm, h Runner) func() {
	w := vault.Bump
	var got func()
	func() {
		defer func() { got, _ = recover().(func()) }()
		h.Run(func() { panic(w) })
	}()
	return got
}

// ThrowBack throws a fresh writer from writable code into the stranger's
// catcher, which throws it on; this call recovers it and runs it.
func ThrowBack(cur realm, h Catcher, shape string) {
	w := vault.Bump
	var v any
	switch shape {
	case "direct":
		v = w
	case "struct":
		v = &WBox{F: w}
	case "slice":
		v = []func(){w}
	case "getter":
		v = func() func() { return w }
	}
	var got any
	func() {
		defer func() { got = recover() }()
		func() {
			defer h.Catch()
			panic(v)
		}()
	}()
	switch x := got.(type) {
	case *WBox:
		x.F()
	case func() func():
		x()()
	default:
		callAny(got)
	}
}

// NestedPrivate makes a writer, makes a nested crossing call, then runs the
// writer it never handed anywhere: that one writes.
func NestedPrivate(cur realm, cb func(cur realm)) {
	w := vault.Bump
	cb(cross(cur))
	w()
}

// NestedShown shows a box holding a writer to a nested crossing call, then
// runs the box's writer.
func NestedShown(cur realm, cb func(cur realm, c *Ctx)) {
	c := &Ctx{Undo: vault.Bump}
	cb(cross(cur), c)
	c.Undo()
}

type GetKeeper interface {
	Keep(f func())
	Pick() func()
}

// DeferKeep hands a fresh writer to a deferred stranger's call that keeps
// it, then runs what the stranger picks.
func DeferKeep(cur realm, h GetKeeper) {
	w := vault.Bump
	func() { defer h.Keep(w) }()
	h.Pick()()
}

type GetThrower interface{ Throw(f func()) }

// DeferThrow hands a fresh writer to a deferred stranger's call that throws
// a getter for it, then runs what the getter returns.
func DeferThrow(cur realm, h GetThrower) {
	w := vault.Bump
	var got any
	func() {
		defer func() { got = recover() }()
		defer h.Throw(w)
	}()
	g, _ := got.(func() func())
	g()()
}

type KeeperMaker interface{ NewKeeper() any }

// LibPick throws a fresh writer from writable code into a keeper the
// stranger's library made, then calls what the keeper picks.
func LibPick(cur realm, h KeeperMaker) {
	k := h.NewKeeper().(Keeper)
	func() {
		defer k.Catch()
		w := vault.Bump
		panic(w)
	}()
	k.Pick()()
}
