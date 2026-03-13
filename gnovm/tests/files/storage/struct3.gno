// PKGPATH: gno.land/r/storage
package storage

var record Foo
var record2 *Foo
var record3 *Foo

func init() {
	NewFoo("a")
}

type Foo struct {
	name string
}

func NewFoo(name string) {
	record = Foo{name}
	record2 = &Foo{name}
}
func SetName(n string) {
	record.name = n
	record2.name = n
}

func GetName() string {
	return record.name + ":" + record2.name + ":" + record3.name
}

func Clear() {
	record = Foo{}
	record2 = nil
}

func main() {
	SetName("aa")
	record3 = record2
	println(GetName())

}

// Output:
// aa:aa:aa

// Storage:
// gno.land/r/storage: 160

// Realm:
// finalizerealm["gno.land/r/storage"]
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:14](6)=
//     @@ -7,14 +7,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "a"
//     +                "value": "aa"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:14",
//              "LastObjectSize": "240",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "RefCount": "1"
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:16](6)=
//     @@ -7,14 +7,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "a"
//     +                "value": "aa"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:16",
//              "LastObjectSize": "241",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15",
//              "RefCount": "1"
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:6](138)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6",
//              "LastObjectSize": "254",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
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
//     +                "Escaped": true,
//     +                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:15](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15",
//     +        "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     -        "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5",
//     -        "RefCount": "1"
//     +        "ModTime": "16",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "1350323698e83f991922ca1d7732a79d01f2161b",
//     +            "Hash": "0a40968d10e2618a057f0c13a02df1150ad975f4",
//                  "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:16"
//              }
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:5](3)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5",
//              "LastObjectSize": "389",
//     -        "ModTime": "13",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "61f2e18dc54206957cd495b6cdc5a945acb9de4a",
//     +                "Escaped": true,
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15"
//                  },
//                  "Index": "0",
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "344",
//     -        "ModTime": "13",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "c2f0254face51f8d660b43c8bbb788e597d32160",
//     +            "Hash": "060a46af87d33d199af9389d47deab6d506bb69c",
//                  "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:14"
//              }
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "fc303c4446c1eb1c7f029ed620c065fb409a4e2a",
//     +                "Hash": "0d15d2e93cd7b0ad2214a9b6a688345652a90e41",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "8ef1494ef5ffb26f92e8e0f996f3b2e3a8a0ca77",
//     +                "Hash": "64212cee5cecb136386bd557bf80fd9797b13943",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5"
//                  }
//              },
//     @@ -54,7 +54,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "71841c6bda7f874839487e182e5e848fe6f5071b",
//     +                "Hash": "be988a2a90e4409bf9fcbab9b610b8d9406dda30",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6"
//                  }
//              },
