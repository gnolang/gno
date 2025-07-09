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
// gno.land/r/storage: 157

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
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15",
//     +        "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +14,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "1350323698e83f991922ca1d7732a79d01f2161b",
//     +            "Hash": "0a40968d10e2618a057f0c13a02df1150ad975f4",
//                  "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:16"
//              }
//          }
