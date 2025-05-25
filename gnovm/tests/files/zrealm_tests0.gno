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
	crossing()

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
// finalizerealm["gno.land/r/demo/tests"]
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:59]={
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
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:59",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58]={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "a7c81f1e07e69c77e7d4cd89ec82fc78cc28fc63",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:59"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57]={
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
//                     "Hash": "e237f5d9b1327cf79b3b81212360b718dab745ed",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:52"
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
//                     "Hash": "06a3eec56308db2a854688299249e9325c8efa05",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:55"
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
//                     "Hash": "e9679398e8552c552254fc5dbfef3a9b34162c9d",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:6",
//         "RefCount": "1"
//     }
// }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:6]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:6",
//     -        "ModTime": "53",
//     +        "ModTime": "56",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//              "RefCount": "1"
//          },
//     @@ -18,11 +18,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "0a4e300c064e340c37a1704afebf583034c7fe96",
//     -                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:54"
//     +                "Hash": "6abf6994a55444381588af249eb4d9185910e7af",
//     +                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57"
//                  },
//     -            "Length": "2",
//     -            "Maxcap": "2",
//     +            "Length": "3",
//     +            "Maxcap": "3",
//                  "Offset": "0"
//              }
//          }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:52]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:52",
//     -        "ModTime": "54",
//     +        "ModTime": "57",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:51",
//              "RefCount": "1"
//          },
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:55]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:55",
//     -        "ModTime": "0",
//     +        "ModTime": "57",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:54",
//              "RefCount": "1"
//          },
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:54]
// finalizerealm["gno.land/r/demo/tests"]
// finalizerealm["gno.land/r/demo/tests"]
// finalizerealm["gno.land/r/demo/tests"]
// finalizerealm["gno.land/r/demo/tests_test"]
