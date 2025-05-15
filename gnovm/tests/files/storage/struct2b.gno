// PKGPATH: gno.land/r/storage_test
package storage_test

type foo struct {
	field string
}

var somevalue foo

func init() {
	somevalue.field = "hello"
}
func main() {
	crossing()
	somevalue.field = "h2"
	print(somevalue)
}

// Output:
// (struct{("h2" string)} gno.land/r/storage_test.foo)



// Storage:
// gno.land/r/storage_test: -3



// Realm:
// finalizerealm["gno.land/r/storage_test"]
// u[334570344df8e77213d2886cb407c37de6da36c1:4](-3)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "hello"
//     +                "value": "h2"
//                  }
//              }
//          ],
