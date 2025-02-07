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
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:19]={
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
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:19",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18]={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6b9b731f6118c2419f23ba57e1481679f17f4a8f",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:19"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17]={
//     "Data": null,
//     "List": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "4"
//             },
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
//                     "Hash": "1cc8ecae9ef4fa531f7e43c9daa8a15b395c7d1b",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "3"
//             },
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
//                     "Hash": "c125144270fc720ab94e98e8ec0a03ec2b617015",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
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
//                     "Hash": "a88c9919a92c42b4405026756fa2bf5bb9a66a9d",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//         "RefCount": "1"
//     }
// }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12]={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12",
//         "ModTime": "17",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "f163acee63433b44deb3168fef87beb588847322",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13"
//         }
//     }
// }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15]={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15",
//         "ModTime": "17",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "2a96c074210eb2db31b7941e31d62deb04e59937",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:16"
//         }
//     }
// }
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests"]
// switchrealm["gno.land/r/demo/tests_test"]
