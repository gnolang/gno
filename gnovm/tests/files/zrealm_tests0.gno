// PKGPATH: gno.land/r/tests/vm0
package tests0

import (
	tests "gno.land/r/tests/vm"
	"gno.land/r/tests/vm/tests_foo"
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
// finalizerealm[tests "gno.land/r/tests/vm"]
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
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:58](363)={
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:58",
//         "LastObjectSize": "363",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:57",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/tests/vm/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5a8bc474b7b9f852ada0457868aa620d418bc175",
//             "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:59"
//         }
//     }
// }
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:57](860)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests/vm/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "64c93b0b09709b4c541ea76b579e2216bbe1938c",
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
//                     "ID": "gno.land/r/tests/vm/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "32edb5532a4c71214c10c036a4c642d3dc0c6416",
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
//                     "ID": "gno.land/r/tests/vm/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "2bb3aa941a5752dd1d2cb986ed7c22fb85e0c30d",
//                     "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:58"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:57",
//         "LastObjectSize": "860",
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
//     -                "Hash": "920c6cdb9efca315c5abdc0f43b6171e4cb02859",
//     -                "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:54"
//     +                "Hash": "dcbf4282f0fef5b96fd3e876df12cb7891438722",
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
//              "LastObjectSize": "368",
//     -        "ModTime": "54",
//     +        "ModTime": "57",
//              "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:51",
//              "RefCount": "1"
//          },
// u[393e33276bb385821437eda0ee7d24f4c0e905fe:55](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:55",
//              "LastObjectSize": "363",
//     -        "ModTime": "0",
//     +        "ModTime": "57",
//              "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:54",
//              "RefCount": "1"
//          },
// d[393e33276bb385821437eda0ee7d24f4c0e905fe:54](-632)
// finalizerealm[tests "gno.land/r/tests/vm"]
// finalizerealm[tests "gno.land/r/tests/vm"]
// finalizerealm[tests "gno.land/r/tests/vm"]
// finalizerealm["gno.land/r/tests/vm0"]
