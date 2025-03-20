// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var f1 = &foo{name: "1"}
var f2 = &foo{name: "2"}
var f3 = &foo{name: "3"}

func main() {
	var Local_S []crossrealm.Fooer
	Local_S = append(Local_S, f1)
	Local_S = append(Local_S, f2) // reference real object, good
	Local_S = append(Local_S, f3)

	fs := Local_S[1:]
	crossrealm.SetSlice(fs) // not crossing realm
	println("ok")
}

// Output:
// hello from crossrealm_test
// hello from crossrealm_test
// ok

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/crossrealm_test"]
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
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "Index": "0",
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
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:5"
//                 },
//                 "Index": "0",
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
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:7"
//                 },
//                 "Index": "0",
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
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "29",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "bf28fe75967956f5280a4373bb55bf679b424f72",
//             "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:5]={
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "IsEscaped": true,
//         "ModTime": "29",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6b8708e350cc6779d2c93481367e979b77c13ce8",
//             "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:6"
//         }
//     }
// }
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:7]={
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:7",
//         "IsEscaped": true,
//         "ModTime": "29",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "453bb359d5162b7bcb3199950e80f3f2889281a8",
//             "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:8"
//         }
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
