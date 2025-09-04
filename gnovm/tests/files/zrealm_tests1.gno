// PKGPATH: gno.land/r/demo/tests
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
// finalizerealm["gno.land/r/demo/tests"]
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9](242)={
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
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9",
//         "LastObjectSize": "242",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8](341)={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8",
//         "LastObjectSize": "341",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests.Foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "763dd26b880a73a0aa811bf6e1f5cb3cb16130e0",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11](244)={
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
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
//         "LastObjectSize": "244",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10](343)={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests.Foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6bf3861029c7e40f75cdf4b1b37cf8439892bd5e",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7](592)={
//     "Data": [],
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.Foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "4f1d4fd770dc5ecb2e9a23fbb044199e313f182f",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8"
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
//                     "ID": "gno.land/r/demo/tests.Foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3eb835a1bc8c1f41b907682ea35b59ff4867a9f0",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7",
//         "LastObjectSize": "592",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3",
//         "RefCount": "1"
//     }
// }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3",
//              "LastObjectSize": "260",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//              "RefCount": "1"
//          },
//     @@ -14,6 +14,17 @@
//                      "ID": "gno.land/r/demo/tests.Stringer"
//                  },
//                  "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.SliceValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "986b16be49c2048ebf2d9eeb7fb7226bbe7e0f7f",
//     +                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
