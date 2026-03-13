// PKGPATH: gno.land/r/storage
package storage

var record Foo
var record2 *Foo
var record3 *Foo

func init() {
	record = Foo{"a"}
	record2 = &Foo{"b"}
}

type Foo struct {
	name string
}

func main() {
	// record is object 5,
	record.name = "aa"
	// record2 is object 7, object 6 is the reference of record 2
	record2.name = "bb"
	record3 = record2

	record = Foo{} // delete object 5
	record2 = nil  //
	println(record3)

}

// Output:
// &(struct{("bb" string)} gno.land/r/storage.Foo)

// Storage:
// gno.land/r/storage: -9

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:13](215)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:13",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:12](6)=
//     @@ -7,14 +7,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "b"
//     +                "value": "bb"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:12",
//              "LastObjectSize": "241",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11",
//              "RefCount": "1"
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:6](135)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6",
//              "LastObjectSize": "254",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,15 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage.Foo"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:11](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11",
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "2b6a1dd38af88a40a05ffd55e77e7ea3b4b3a156",
//     +            "Hash": "6c3ebd5509617364a76aada1372a2c3f2afa0dc4",
//                  "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:12"
//              }
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "344",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "332e3d0f721fc783bbd5d5fec75baae6488cad5a",
//     -            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:10"
//     +            "Hash": "95a698670cebb93239a3707db3d217c4799691dd",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:13"
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:5](-130)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5",
//              "LastObjectSize": "389",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,16 +13,6 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage.Foo"
//                  }
//     -        },
//     -        "V": {
//     -            "@type": "/gno.PointerValue",
//     -            "Base": {
//     -                "@type": "/gno.RefValue",
//     -                "Hash": "ca7da93d48d5cf7282da3ef614d9a6718a9a9b4f",
//     -                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11"
//     -            },
//     -            "Index": "0",
//     -            "TV": null
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1090",
//     -        "ModTime": "12",
//     +        "ModTime": "13",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "012dad3c2e41fe815f2d9a094f137fb964f34b02",
//     +                "Hash": "d1292000b478f5abe3e0be95d29406bb56ca5ecf",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "6c1bcf70ea66520a1c804fbda027e3589ab7b8c1",
//     +                "Hash": "99f226279aeecaf7b03c427936bd26ad127e9bac",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5"
//                  }
//              },
//     @@ -54,7 +54,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "71841c6bda7f874839487e182e5e848fe6f5071b",
//     +                "Hash": "ade5ad0f19effc655b9075728d48d3618f72c671",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6"
//                  }
//              },
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:10](-240)
