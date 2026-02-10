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
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:10](-240)
