// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
	b    *bar
}
type bar struct{ name string }

func (fooer) Foo() { println("hello") }

var f = fooer{name: "local_fooer"} // already real
var b = &bar{name: "local_bar"}

func main() {
	f.b = b                      // this makes f dirty, it's real
	crossrealm.SetContainer2(&f) // object DidUpdate is package block
	crossrealm.RunF()
	print(".")
}

// Output:
// hello
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/iface"]
// u[r:5836d1a3cec217737a5553d4c1a03d8897d5f281:4]={
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
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2"
//                 },
//                 "Index": "2",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:5836d1a3cec217737a5553d4c1a03d8897d5f281:4",
//         "ModTime": "4",
//         "OwnerID": "r:5836d1a3cec217737a5553d4c1a03d8897d5f281:2",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "local_fooer"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.bar"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "ModTime": "6",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
// }
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.bar"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5b99c134e6faa2e6333fe3f0e7fc480ea5d493d2",
//             "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:5"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm/iface"]
// switchrealm["gno.land/r/crossrealm_test"]
