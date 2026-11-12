// PKGPATH: gno.land/r/test
package test

// Each case leaves a reference cycle with no escaped node at finalize, after
// a persisted object is deleted in the same transaction; the save walk must
// still terminate.

type Node struct{ Next *Node }

var (
	Root *Node
	M    map[string]any
	R    *Node
)

func init() {
	Root = &Node{}
	M = map[string]any{"k": 1}
	R = &Node{}
}

func main(cur realm) {
	// a new self-cycle loses its escape mark when Root is deleted.
	x := &Node{}
	x.Next = x
	Root.Next = x
	Root = nil

	// the same, as a map that contains itself.
	m := map[string]any{}
	m["self"] = m
	M["x"] = m
	M = nil

	// a deleted persisted object revived by a new object it points back to.
	p := R
	R = nil
	n := &Node{Next: p}
	p.Next = n

	println("ok")
}

// Output:
// ok

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:15](382)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15",
//         "LastObjectSize": "382",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:14](340)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//         "IsEscaped": true,
//         "LastObjectSize": "340",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "7879a4edb94d3a90114f6f5eaa16748578e10c77",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15"
//         }
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:16](433)={
//     "List": {
//         "List": [
//             {
//                 "Key": {
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "16"
//                     },
//                     "V": {
//                         "@type": "/gno.StringValue",
//                         "value": "self"
//                     }
//                 },
//                 "Value": {
//                     "T": {
//                         "@type": "/gno.MapType",
//                         "Key": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "16"
//                         },
//                         "Value": {
//                             "@type": "/gno.InterfaceType",
//                             "Generic": "",
//                             "Methods": [],
//                             "PkgPath": ""
//                         }
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:16"
//                     }
//                 }
//             }
//         ]
//     },
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:16",
//         "IsEscaped": true,
//         "LastObjectSize": "433",
//         "ModTime": "0",
//         "RefCount": "1"
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:13](135)=
//     @@ -7,13 +7,22 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//              "LastObjectSize": "252",
//     -        "ModTime": "0",
//     +        "ModTime": "13",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:12](6)=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "LastObjectSize": "338",
//     -        "ModTime": "0",
//     -        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//     +        "ModTime": "18",
//     +        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18",
//              "RefCount": "1"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "7382660a24f6a8370e19e5f2a49230f6b11d6755",
//     +            "Hash": "04beebb95cbd64780e0ee0ef0eea44ec7d73dc07",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//              }
//          }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:18](382)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "53d688bdf751e11b0501e3d726f1397cc933e1af",
//                     "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18",
//         "LastObjectSize": "382",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:17](340)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17",
//         "IsEscaped": true,
//         "LastObjectSize": "340",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3b7bb166c720ba781632c186b7af78605b16f35b",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](-129)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "386",
//     -        "ModTime": "8",
//     +        "ModTime": "13",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,16 +13,6 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.Node"
//                  }
//     -        },
//     -        "V": {
//     -            "@type": "/gno.PointerValue",
//     -            "Base": {
//     -                "@type": "/gno.RefValue",
//     -                "Hash": "6ee1084626e5131663e5e6b611b8254532545b52",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//     -            },
//     -            "Index": "0",
//     -            "TV": null
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](-107)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "368",
//     -        "ModTime": "8",
//     +        "ModTime": "13",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,11 +19,6 @@
//                      "Methods": null,
//                      "PkgPath": ""
//                  }
//     -        },
//     -        "V": {
//     -            "@type": "/gno.RefValue",
//     -            "Hash": "b35308ae288685e73f84d9b0cf826e02ddf94964",
//     -            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:5](-130)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "LastObjectSize": "387",
//     -        "ModTime": "8",
//     +        "ModTime": "13",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,16 +13,6 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.Node"
//                  }
//     -        },
//     -        "V": {
//     -            "@type": "/gno.PointerValue",
//     -            "Base": {
//     -                "@type": "/gno.RefValue",
//     -                "Hash": "1b8ceb52739a85426fc3fb538e6a32b0ac19c0ac",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//     -            },
//     -            "Index": "0",
//     -            "TV": null
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1044",
//     -        "ModTime": "13",
//     +        "ModTime": "18",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d4ccebc5cc04e5aa7e78c2a6fdff3b44a5c45981",
//     +                "Hash": "96eb987936a8eee56546aef3bc9b5871cd2f4bd0",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "8084d99a1b11344dd85d927863fb2c0f818e4439",
//     +                "Hash": "f3494a9e2fa18ae02075aae7dba64fecbeb0a898",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
//     @@ -66,7 +66,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "a4f764a23a2b0fbacbad7ee6948cbc4caa9d0f11",
//     +                "Hash": "4c119c887268831feb0e2de9c28b63efdea7dfa6",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:9](-337)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:10](-251)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:11](-279)
