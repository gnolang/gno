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
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:19](245)={
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
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18](356)={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17",
//         "RefCount": "1"
//     },
//     "Value": {
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
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17](839)={
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
//                     "Hash": "4e0d77a91ba35733bf82329317bf8c8dffa6f655",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12"
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
//                     "Hash": "fa414e1770821b8deb8e6d46d97828c47f7d5fa5",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15"
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
//                     "Hash": "aaa64d049cf8660d689780acac9f546f270eaa4e",
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
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2](0)=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//              "IsEscaped": true,
//     -        "ModTime": "13",
//     +        "ModTime": "16",
//              "RefCount": "5"
//          },
//          "Parent": null,
//     @@ -73,11 +73,11 @@
//                      "@type": "/gno.SliceValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "694cb4385eb8d5e25fe7737cbc45695c7ac83f3c",
//     -                    "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14"
//     -                },
//     -                "Length": "2",
//     -                "Maxcap": "2",
//     +                    "Hash": "c2aab89570a9faa4acf463a140a04e9b96fae73d",
//     +                    "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17"
//     +                },
//     +                "Length": "3",
//     +                "Maxcap": "3",
//                      "Offset": "0"
//                  }
//              },
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12](0)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12",
//     -        "ModTime": "14",
//     +        "ModTime": "17",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
//              "RefCount": "1"
//          },
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15](2)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:15",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14",
//              "RefCount": "1"
//          },
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14](-618)
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests"]
// switchrealm["gno.land/r/demo/tests_test"]
