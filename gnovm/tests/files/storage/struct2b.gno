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
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:4](-3)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "hello"
//     +                "value": "h2"
//                  }
//              }
//          ],
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "d0f1ba7ee33fbbbf66a1c14967ed5c48b66e1bca",
//     +            "Hash": "9db24be033570705002053c5abb31699e4f69c24",
//                  "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:4"
//              }
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5bda02c49358fcee26ac0216e11955619a59e04c",
//     +                "Hash": "c67a93ed3f5518de5e794de0afd125d9e54ccc16",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
