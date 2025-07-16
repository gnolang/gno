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
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:13:0](219)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:13:0",
//         "LastObjectSize": "219",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//         "RefCount": "1"
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:12:0](6)=
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
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:12:0",
//              "LastObjectSize": "245",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11:0",
//              "RefCount": "1"
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:11:0](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11:0",
//              "LastObjectSize": "346",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5:0",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "8a17dbb0b074808d688a3a04cb464576e3965c78",
//     +            "Hash": "9a8d88a4c46e72a2ba5763c33fbe1e192b2b7eee",
//                  "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:12:0"
//              }
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:6:0](138)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6:0",
//              "LastObjectSize": "258",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage.Foo"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "92a5c3785383d06455901480b85f801c2327d3f9",
//     +                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11:0"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3:0](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//              "LastObjectSize": "350",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "b799280fb2331fc7850ccc95887a77e29931e577",
//     -            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:10:0"
//     +            "Hash": "82537e190a338ad659ecc4c659fcf10ce6ba86cc",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:13:0"
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:5:0](-133)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5:0",
//              "LastObjectSize": "396",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2:0",
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
//     -                "Hash": "dbbe00e8345cdc3a733f81bc89bcfaef6dedd7e1",
//     -                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:11:0"
//     -            },
//     -            "Index": "0",
//     -            "TV": null
//              }
//          }
//      }
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:10:0](-244)
