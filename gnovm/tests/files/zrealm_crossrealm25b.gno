// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	name string
}

func (fooer) Foo() { println("hello " + std.CurrentRealm().PkgPath()) }

type localContainer struct {
	name string
	f    fooer
}

var lc localContainer

func init() {
	// container attached, embedded object attached
	lc = localContainer{name: "local_container", f: fooer{name: "local_fooer"}} // attach
}

func main() {
	crossrealm.SetContainer(&lc.f)
	print(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:8]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.fooer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
//                 },
//                 "Index": "1",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:8",
//         "ModTime": "44",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
