// PKGPATH: gno.land/r/zrealm_chalc_control_param
package zrealm_chalc_control_param

var count int

func attack() { run(func() { count++ }) }

// Control: the parameter alone, no := reuse.
func run(g func()) {
	g()
}

func main(cur realm) {
	println("before:", count)
	attack()
	println("after:", count)
}

// Output:
// before: 0
// after: 1
