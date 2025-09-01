// PKGPATH: gno.land/r/tests/tests
package tests

type Stringer interface {
	String() string
}

type Foo struct{ name string }

func (f *Foo) String() string {
	return "..."
}

var stringers []Stringer

func AddStringer(str Stringer) {
	stringers = append(stringers, str)
}

func main() {
	f1 := &Foo{name: "foo1"}
	AddStringer(f1)
	f2 := &Foo{name: "foo2"}
	AddStringer(f2)
	println("ok")
}

// Output:
// ok

// Realm:
// finalizerealm["gno.land/r/tests/tests"]
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:9](242)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "foo1"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:9",
//         "LastObjectSize": "242",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:8",
//         "RefCount": "1"
//     }
// }
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:8](342)={
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:8",
//         "LastObjectSize": "342",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:7",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/tests/tests.Foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "32b0cd931fbf65aae35982872a5e471d6cb27a27",
//             "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:9"
//         }
//     }
// }
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:11](244)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "foo2"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:11",
//         "LastObjectSize": "244",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:10",
//         "RefCount": "1"
//     }
// }
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:10](344)={
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:10",
//         "LastObjectSize": "344",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:7",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/tests/tests.Foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6d92e39b0061a0613b0d325846aa75bbb691803a",
//             "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:11"
//         }
//     }
// }
// c[393e33276bb385821437eda0ee7d24f4c0e905fe:7](594)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests/tests.Foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "80ed8924d7a64098d715f02f058df39828ba4724",
//                     "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:8"
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
//                     "ID": "gno.land/r/tests/tests.Foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "40bb80e2ce67289e78ceff2e0f7d8ef752c33dd2",
//                     "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:10"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:7",
//         "LastObjectSize": "594",
//         "ModTime": "0",
//         "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:3",
//         "RefCount": "1"
//     }
// }
// u[393e33276bb385821437eda0ee7d24f4c0e905fe:3](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "393e33276bb385821437eda0ee7d24f4c0e905fe:3",
//              "LastObjectSize": "261",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "393e33276bb385821437eda0ee7d24f4c0e905fe:2",
//              "RefCount": "1"
//          },
//     @@ -14,6 +14,17 @@
//                      "ID": "gno.land/r/tests/tests.Stringer"
//                  },
//                  "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.SliceValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "c6eb246ed16f7c6f4ed6d0d30eb264316d578502",
//     +                "ObjectID": "393e33276bb385821437eda0ee7d24f4c0e905fe:7"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
