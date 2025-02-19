// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type container struct {
	name string
	f    *fooer
}

func (container) Foo() { println("hello " + c.f.name) }

type fooer struct{ name string }

var c = &container{name: "local_container"}

func main() {
	ff := &fooer{name: "local_foo"}
	c.f = ff
	crossrealm.SetFooer(c)
	crossrealm.CallFooerFoo()
	print(".")
}

// Output:
// hello local_foo
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "5",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.container"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:7]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "local_foo"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:7",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "RefCount": "1"
//     }
// }
// c[f5a516808f8976c33939133293d598ce3bca4e8d:6]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.fooer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "90e6414435c56b7ffa17bd575535b7a558b93e22",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:7"
//         }
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "local_container"
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
//                     "Hash": "7d3bda685439b3989f18e824762f473f40c944fb",
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "5",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "RefCount": "1"
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "7",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.container"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "ef7d0f54195eff59079f178206ba725d07f095a7",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
