// PKGPATH: gno.land/r/demo/reflectkinds
package reflectkinds

// Which values get an address of their own, and which borrow nothing.
//
// The rule is one line: a value has an address when it is its own object, and
// not when its identity would really be somebody else's. Everything below is
// that rule applied.

import (
	"chain/reflect"
)

type inner struct{ n int }

type outer struct {
	field inner
	arr   [2]inner
}

var (
	standalone = &inner{n: 1}
	container  = &outer{}
	callback   = func() int { return 1 }
	numbers    = []int{1, 2, 3}
	nilPtr     *inner
)

func show(label string, v any) {
	obj, ok := reflect.Of(v)
	switch {
	case ok:
		println(label, "->", obj.Address())
	case reflect.HasIdentity(v):
		println(label, "-> (an object, not persisted yet)")
	default:
		println(label, "-> (no identity of its own)")
	}
}

func main(cur realm) {
	// Its own object. `&T{...}` and `new(T)` produce these.
	show("&inner{}        ", standalone)
	show("&outer{}        ", container)

	// A func value is a reference: two variables holding one closure hold the
	// same object, so a stored callback is addressable separately from
	// whatever holds it.
	show("func() int      ", callback)

	// A pointer into a container is not its own object. Were &container.field
	// to report the container's address, container.arr[0] would report the
	// same one, and a single address would name three different things.
	show("&container.field", &container.field)
	show("&container.arr[0]", &container.arr[0])

	// A slice resolves to its backing array, which every other view of it
	// shares, so numbers[0:1] and numbers[1:2] would agree.
	show("[]int           ", numbers)

	// No object behind them at all.
	show("42              ", 42)
	show(`"hello"         `, "hello")
	show("(*inner)(nil)   ", nilPtr)

	println("")

	// A value created by this call is an object with no address yet, which is
	// a different answer from having none. show prints them differently above;
	// here is the check a realm would write.
	fresh := &inner{n: 9}
	_, ok := reflect.Of(fresh)
	println("fresh: addressable now:", ok, "/ ever:", reflect.HasIdentity(fresh))

	// Nothing without an identity ever yields a non-zero Object, so a caller
	// that forgets to check ok collapses them all onto the zero value rather
	// than inventing addresses.
	for _, v := range []any{&container.field, numbers, 42, nilPtr} {
		if obj, ok := reflect.Of(v); ok || !obj.IsZero() || obj.Address() != "" {
			println("LEAKED AN ADDRESS")
		}
	}
	println("no addresses leaked from values without identity")

	println(".")
}

// Output:
// &inner{}         -> g1aecuuaguc5654vk3yaa2jzlsr36tu543gfjuky
// &outer{}         -> g185ejn3r7w6gef5yk6mv8xj4g34d9s57telt7h9
// func() int       -> g1tgzz9nkvgtlr90cgha8g8lajvn9p2pew5epgmd
// &container.field -> (no identity of its own)
// &container.arr[0] -> (no identity of its own)
// []int            -> (no identity of its own)
// 42               -> (no identity of its own)
// "hello"          -> (no identity of its own)
// (*inner)(nil)    -> (no identity of its own)
//
// fresh: addressable now: false / ever: true
// no addresses leaked from values without identity
// .
