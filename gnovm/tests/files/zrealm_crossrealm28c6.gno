// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var arr [3]*foo // arr and all elems are all real
var s []*foo

func init() {
	var f1 = foo{name: "1"}
	var f2 = foo{name: "2"}
	var f3 = foo{name: "3"}
	arr[0] = &f1
	arr[1] = &f2
	arr[2] = &f3
}

func main() {
	s = arr[:]
	println("ok")
}

// Output:
// ok

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:3]={
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
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:5"
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
//                     "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:5"
//                 },
//                 "Index": "2",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "8",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     }
// }
