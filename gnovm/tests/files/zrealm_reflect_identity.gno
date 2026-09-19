// PKGPATH: gno.land/r/demo/reflectkinds
package reflectkinds

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
	numbers    = []int{1, 2, 3}
	nilPtr     *inner
)

func main(cur realm) {
	// A pointer to a standalone object: the case the package exists for.
	// `&T{...}` gives the object its own heap item, so it has its own ID.
	println("standalone:", reflect.HasIdentity(standalone))

	// A pointer into a container borrows nothing. Were &container.field to
	// report the container's ID, container.arr[0] would report the same one,
	// and one ID would name three different things.
	println("struct field:", reflect.HasIdentity(&container.field))
	println("array element:", reflect.HasIdentity(&container.arr[0]))

	// A slice resolves to its backing array, which every other view of it
	// shares, so numbers[0:1] and numbers[1:2] would agree.
	println("slice:", reflect.HasIdentity(numbers))

	// Values with no object behind them at all.
	println("scalar:", reflect.HasIdentity(42))
	println("string:", reflect.HasIdentity("hello"))
	println("nil pointer:", reflect.HasIdentity(nilPtr))

	// None of those aborts the call, and none of them yields an ID: a caller
	// checking ok gets one uniform answer instead of a mix of empty strings
	// and panics.
	for _, v := range []interface{}{&container.field, numbers, 42, nilPtr} {
		if id, ok := reflect.ObjectIDOf(v); ok || !id.IsZero() {
			println("LEAKED AN ID")
		}
	}
	println("no IDs leaked from values without identity")

	// The container itself does have one, being standalone.
	containerID, containerOK := reflect.ObjectIDOf(container)
	standaloneID, standaloneOK := reflect.ObjectIDOf(standalone)
	println("container has an ID:", containerOK)
	println("standalone has an ID:", standaloneOK)
	println("and they differ:", containerID != standaloneID)

	println(".")
}

// Output:
// standalone: true
// struct field: false
// array element: false
// slice: false
// scalar: false
// string: false
// nil pointer: false
// no IDs leaked from values without identity
// container has an ID: true
// standalone has an ID: true
// and they differ: true
// .
