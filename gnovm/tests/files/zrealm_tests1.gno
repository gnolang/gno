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
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7](383)={
//     "Data": null,
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
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7",
//         "LastObjectSize": "383",
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
//     +                "Hash": "4cff64924c0ff790bfd15a63aa8936b06f688388",
//     +                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7"
//     +            },
//     +            "Length": "1",
//     +            "Maxcap": "1",
//     +            "Offset": "0"
//              }
//          }
//      }
// finalizerealm["gno.land/r/demo/tests"]
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8",
//              "LastObjectSize": "341",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7",
//              "RefCount": "1"
//          },
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12](244)={
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
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12",
//         "LastObjectSize": "244",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11](344)={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
//         "LastObjectSize": "344",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests.Foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "e010d5dccd70779f311cc7714f40303eb4e825e0",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10](593)={
//     "Data": null,
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
//                     "Hash": "b63bfcb13075d041fabde710e4e31dfe7a42e781",
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
//                     "Hash": "433e61840bc578153a509010dbb5b3ab785256db",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10",
//         "LastObjectSize": "593",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3",
//         "RefCount": "1"
//     }
// }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3](1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3",
//              "LastObjectSize": "397",
//     -        "ModTime": "6",
//     +        "ModTime": "9",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//              "RefCount": "1"
//          },
//     @@ -19,11 +19,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "4cff64924c0ff790bfd15a63aa8936b06f688388",
//     -                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7"
//     +                "Hash": "7b7d4e2dc4385356c50e309b2263e8ab059959ca",
//     +                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:10"
//                  },
//     -            "Length": "1",
//     -            "Maxcap": "1",
//     +            "Length": "2",
//     +            "Maxcap": "2",
//                  "Offset": "0"
//              }
//          }
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7](-383)
// finalizerealm["gno.land/r/demo/tests"]
