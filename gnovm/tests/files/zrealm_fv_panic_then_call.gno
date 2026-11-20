// PKGPATH: gno.land/r/fvpanicthencall
package fvpanicthencall

import "gno.land/r/tests/vm/fvx"

func try(p *fvx.T) (n int) {
	defer func() {
		if r := recover(); r != nil {
			n = -1
		}
	}()
	return p.Add(div(0))
}

func div(z int) int { return 10 / z }

func main(cur realm) {
	p := fvx.Ptr()
	println(try(p))
	println(p.Val(), p.Add(1))
	defer func() { println("deferred closure", p.Val()) }()
}

// Output:
// -1
// 1 2
// deferred closure 1
