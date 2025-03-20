// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(func() crossrealm.Fooer { return f })
	crossrealm.CallFooerGetterFoo()
	f.SetS("B")
	crossrealm.CallFooerGetterFoo()
	println(".")

	f.SetS("C")
	crossrealm.SetFooerGetter(crossrealm_b.FooerGetter)
	crossrealm.CallFooerGetterFoo()
	println(".")

	f.SetS("D")
	crossrealm.SetFooerGetter(crossrealm_b.FooerGetterBuilder())
	crossrealm.CallFooerGetterFoo()
	println(".")
}

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello B cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
// hello C cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
// hello D cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// u[r:0edc46caf30c00efd87b6c272673239eafbd051e:3]={
//     "ObjectInfo": {
//         "ID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:3",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "OwnerID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests/crossrealm_b.fooer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "75480330ecbe91b42c19cb63b3e8908d4ea9900e",
//             "ObjectID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:4"
//         }
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[r:0edc46caf30c00efd87b6c272673239eafbd051e:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "B"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:4",
//         "ModTime": "5",
//         "OwnerID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:3",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[r:0edc46caf30c00efd87b6c272673239eafbd051e:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "C"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:4",
//         "ModTime": "5",
//         "OwnerID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:3",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// u[r:0edc46caf30c00efd87b6c272673239eafbd051e:3]={
//     "ObjectInfo": {
//         "ID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:3",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "OwnerID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests/crossrealm_b.fooer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "05f23d7eb396436367d0e5c55f21da75458cc082",
//             "ObjectID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:4"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[r:0edc46caf30c00efd87b6c272673239eafbd051e:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "D"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:4",
//         "ModTime": "5",
//         "OwnerID": "r:0edc46caf30c00efd87b6c272673239eafbd051e:3",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
