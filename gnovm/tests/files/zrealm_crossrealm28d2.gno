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
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:7"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "1"
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "29",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5773aa703bad7f19a0b46088a3dee0e05264bc05",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:5]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "IsEscaped": true,
//         "ModTime": "29",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "c8796d0727128e5790ad98d57654f528404d8508",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
//         }
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:7]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:7",
//         "IsEscaped": true,
//         "ModTime": "29",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "4336baa01760651f4daf8ddf42c3d8b8db48161b",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:8"
//         }
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
