// PKGPATH: gno.land/r/tests0
package tests0

import (
	"gno.land/r/tests"
	"gno.land/r/tests_foo"
)

func init() {
	tests_foo.AddFooStringer("one")
	tests_foo.AddFooStringer("two")
}

func main(cur realm) {
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
// finalizerealm["gno.land/r/tests"]
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:59](245)={
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
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:59",
//         "LastObjectSize": "245",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:58",
//         "RefCount": "1"
//     }
// }
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:58](357)={
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:58",
//         "LastObjectSize": "357",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:57",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5a8bc474b7b9f852ada0457868aa620d418bc175",
//             "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:59"
//         }
//     }
// }
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:57](842)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "8fd796ea9c48176b0992aa3e0fb833dafcaa313d",
//                     "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:52"
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
//                     "ID": "gno.land/r/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "ef0b89c691f304e747850c40ec0e8531da38f77d",
//                     "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:55"
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
//                     "ID": "gno.land/r/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "e20b5048f56207acfcaef5875a07ed09419243a5",
//                     "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:58"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:57",
//         "LastObjectSize": "842",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:6",
//         "RefCount": "1"
//     }
// }
// u[393e33276bb385821437eda0ee7d24f4c0e905fe:6](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:6",
//              "LastObjectSize": "399",
//     -        "ModTime": "53",
//     +        "ModTime": "56",
//              "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:2",
//              "RefCount": "1"
//          },
//     @@ -19,11 +19,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c100efd2b64b5fac2f0ce962d2cb4f645b30a7ae",
//     -                "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:54"
//     +                "Hash": "3a310d8500821e76919885ffa8707ef0696293d7",
//     +                "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:57"
//                  },
//     -            "Length": "2",
//     -            "Maxcap": "2",
//     +            "Length": "3",
//     +            "Maxcap": "3",
//                  "Offset": "0"
//              }
//          }
// u[393e33276bb385821437eda0ee7d24f4c0e905fe:52](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:52",
//              "LastObjectSize": "362",
//     -        "ModTime": "54",
//     +        "ModTime": "57",
//              "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:51",
//              "RefCount": "1"
//          },
// u[393e33276bb385821437eda0ee7d24f4c0e905fe:55](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:55",
//              "LastObjectSize": "357",
//     -        "ModTime": "0",
//     +        "ModTime": "57",
//              "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:54",
//              "RefCount": "1"
//          },
// d[393e33276bb385821437eda0ee7d24f4c0e905fe:54](-620)
// finalizerealm["gno.land/r/tests"]
// finalizerealm["gno.land/r/tests"]
// finalizerealm["gno.land/r/tests"]
// finalizerealm["gno.land/r/tests0"]
