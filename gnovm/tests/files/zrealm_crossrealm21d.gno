// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
	bar
}
type bar struct{ name string }

func (fooer) Foo() { println("hello") }

var f = fooer{name: "local_fooer"}

func main() {
	f.bar = bar{name: "local_bar"} // this makes f dirty, it's real
	crossrealm.SetContainer2(&f)   // object DidUpdate is package block
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
// c[f5a516808f8976c33939133293d598ce3bca4e8d:6]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "local_bar"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "RefCount": "1"
//     }
// }
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
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.bar"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "dac9452dd14f52ea03ff0d5898b481c8e993a412",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "ModTime": "5",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
// }
// d[f5a516808f8976c33939133293d598ce3bca4e8d:4]
// switchrealm["gno.land/r/demo/tests/crossrealm/iface"]
// switchrealm["gno.land/r/crossrealm_test"]
