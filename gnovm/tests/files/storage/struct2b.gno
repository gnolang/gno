// PKGPATH: gno.land/r/storage
package storage

type foo struct {
	field string
}

var somevalue foo

func init() {
	somevalue.field = "hello"
}
func main() {
	somevalue.field = "h2"
	print(somevalue)
}

// Output:
// (struct{("h2" string)} gno.land/r/storage.foo)

// Storage:
// gno.land/r/storage: -3

// Realm:
// finalizerealm["gno.land/r/storage"]
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:4](-3)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "hello"
//     +                "value": "h2"
//                  }
//              }
//          ],
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "cb71b821cefa1bb05605966f33f2d53001f59237",
//     +            "Hash": "f5d337909899c4471d6bec0860fc203738d89be4",
//                  "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:4"
//              }
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:2](0)=
//     @@ -78,7 +78,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5122c40762af4d008a4f359648d1613fe38cd2d6",
//     +                "Hash": "05facf5d1289d68225e98ccea855944a633bd134",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
