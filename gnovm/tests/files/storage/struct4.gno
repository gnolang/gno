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
// c[04f997a9ca158338c03cfc00686d77220a6cf62f:13](215)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:13",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     }
// }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:12](6)=
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
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:12",
//              "LastObjectSize": "241",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:11",
//              "RefCount": "1"
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:6](135)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:6",
//              "LastObjectSize": "254",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
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
//     +                "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:11"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:11](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:11",
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:5",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "51c6fa6364690a448a418ec9a08004b1bb4c6769",
//     +            "Hash": "496985af13c208547556c56c1148bcd7bda561ef",
//                  "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:12"
//              }
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "344",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "04e1534ca648348a54ba752206271062e0c1496a",
//     -            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:10"
//     +            "Hash": "326d056615debb3591e04b2be40e3f29bcfb4701",
//     +            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:13"
//              }
//          }
//      }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:5](-130)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:5",
//              "LastObjectSize": "389",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
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
//     -                "Hash": "1e9529eaabb65e7bdcac367f5fe455f6ef66f98c",
//     -                "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:11"
//     -            },
//     -            "Index": "0",
//     -            "TV": null
//              }
//          }
//      }
// d[04f997a9ca158338c03cfc00686d77220a6cf62f:10](-240)
