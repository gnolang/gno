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
// u[5836d1a3cec217737a5553d4c1a03d8897d5f281:4]={
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:2"
//                 },
//                 "Index": "2",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "5836d1a3cec217737a5553d4c1a03d8897d5f281:4",
//         "ModTime": "4",
//         "OwnerID": "5836d1a3cec217737a5553d4c1a03d8897d5f281:2",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "ModTime": "6",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.bar"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6d19add23dca2c83b923c7d597cefd31a2d5bd90",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm/iface"]
// switchrealm["gno.land/r/crossrealm_test"]
