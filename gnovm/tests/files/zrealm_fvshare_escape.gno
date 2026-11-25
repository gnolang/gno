// PKGPATH: gno.land/r/zrealm_fvshare_escape
// A writer captured by a stored closure, or held as a bound method's
// receiver, must not leave the read-only run by panic or by a captured
// variable and keep writing.
package zrealm_fvshare_escape

type F func()

func (f F) Throw()  { panic(f) }
func (f F) Get() F  { return f }

var (
	count int
	G     func()
	GR    func() func()
	BT    func()
	BG    func() F
)

func catch(g func()) (r any) {
	defer func() { r = recover() }()
	g()
	return nil
}

func try(name string, f func()) {
	defer func() {
		r := recover()
		println(name, "refused:", r != nil, "count:", count)
	}()
	f()
}

func main(cur realm) {
	w := func() { count++ }
	G = func() { panic(w) }
	GR = func() func() { return w }
	try("stored copy called", func() { G() })
	try("returned from stored", func() { GR()() })
	r := catch(G)
	try("panicked from stored", r.(func()))

	var out func()
	G = func() { out = w }
	try("stored copy writes capture", func() { G() })
	try("out after stored write", func() {
		if out == nil {
			panic("out not set")
		}
		out()
	})
	G = nil // drop the stored closure so out is not saved

	v := func() { count++ }
	BT = F(v).Throw
	BG = F(v).Get
	try("bound get", func() { BG()() })
	r2 := catch(BT)
	try("bound throw", r2.(F))
	println("done")
}

// Output:
// stored copy called refused: true count: 0
// returned from stored refused: true count: 0
// panicked from stored refused: true count: 0
// stored copy writes capture refused: false count: 0
// out after stored write refused: true count: 0
// bound get refused: true count: 0
// bound throw refused: true count: 0
// done
