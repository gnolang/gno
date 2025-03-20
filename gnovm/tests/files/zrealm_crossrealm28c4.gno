// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var f1 = foo{name: "1"}
var f2 = foo{name: "2"}
var f3 = foo{name: "3"}

func main() {
	var arr [3]crossrealm.Fooer // unreal array
	arr[0] = &f1
	arr[1] = &f2
	arr[2] = &f3

	crossrealm.SetArray(arr)
	println("ok")
}

// Output:
// &(struct{("1" string)} gno.land/r/crossrealm_test.foo)
// &(struct{("2" string)} gno.land/r/crossrealm_test.foo)
// &(struct{("3" string)} gno.land/r/crossrealm_test.foo)
// ok

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/slice"]
// c[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2"
//                 },
//                 "Index": "1",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2"
//                 },
//                 "Index": "2",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2"
//                 },
//                 "Index": "3",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "ModTime": "0",
//         "OwnerID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "1"
//     }
// }
// d[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:4]
// switchrealm["gno.land/r/crossrealm_test"]
