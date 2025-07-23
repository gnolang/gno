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
