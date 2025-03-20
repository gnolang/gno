// PKGPATH: gno.land/r/demo/tests_test
package tests_test

import (
	"gno.land/r/demo/tests"
	"gno.land/r/demo/tests_foo"
)

func init() {
	tests_foo.AddFooStringer("one")
	tests_foo.AddFooStringer("two")
}

func main() {
	tests_foo.AddFooStringer("three")
	println(tests.Render(""))
	println("end")
}

// Output:
// 0: &FooStringer{one}
// 1: &FooStringer{two}
// 2: &FooStringer{three}
//
// end

// Realm:
// switchrealm["gno.land/r/demo/tests"]
// c[r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:19]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "three"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:19",
//         "ModTime": "0",
//         "OwnerID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18",
//         "RefCount": "1"
//     }
// }
// c[r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18]={
//     "ObjectInfo": {
//         "ID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18",
//         "ModTime": "0",
//         "OwnerID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "93c8b678657f67ff13ac39b31d8db778d96695e5",
//             "ObjectID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:19"
//         }
//     }
// }
// c[r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "b421f4333ca7eee6058e0740311145e4f725c6df",
//                     "ObjectID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12"
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
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "c41afec0e0214d8c10709fb43273d731e81b7559",
//                     "ObjectID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15"
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
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "a5b168f123f4f641086e0b9cb319860a8fb330ac",
//                     "ObjectID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17",
//         "ModTime": "0",
//         "OwnerID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//         "RefCount": "1"
//     }
// }
// u[r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12]={
//     "ObjectInfo": {
//         "ID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12",
//         "ModTime": "17",
//         "OwnerID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "712ead1805462af46e84473fe440faec02d9e37d",
//             "ObjectID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13"
//         }
//     }
// }
// u[r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15]={
//     "ObjectInfo": {
//         "ID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15",
//         "ModTime": "17",
//         "OwnerID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "0e8a951402efc97b905432060a6cf80124c6ff29",
//             "ObjectID": "r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:16"
//         }
//     }
// }
// d[r:0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests"]
// switchrealm["gno.land/r/demo/tests_test"]
