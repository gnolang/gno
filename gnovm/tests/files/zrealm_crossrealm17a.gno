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
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "5",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.container"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// c[r:f5a516808f8976c33939133293d598ce3bca4e8d:7]={
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
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:7",
//         "ModTime": "0",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "RefCount": "1"
//     }
// }
// c[r:f5a516808f8976c33939133293d598ce3bca4e8d:6]={
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "ModTime": "0",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.fooer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "35dcda6e017a0d99077a159c19fca8b9f729dd67",
//             "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:7"
//         }
//     }
// }
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:4]={
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
//                     "Hash": "4144d0c59ab9bf408f9f66b6d73ccdb052dd4478",
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:6"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "5",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
