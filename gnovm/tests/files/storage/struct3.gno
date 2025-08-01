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
// gno.land/r/storage: 159

// Realm:
// finalizerealm["gno.land/r/storage"]
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:14:0](6)=
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
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:14:0",
//              "LastObjectSize": "244",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//              "RefCount": "1"
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:16:0](6)=
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
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:16:0",
//              "LastObjectSize": "245",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15:0",
//              "RefCount": "1"
//          }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:6:0](140)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6:0",
//              "LastObjectSize": "258",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
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
//     +                "Escaped": true,
//     +                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15:0"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:15:0](7)=
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:15:0",
//     +        "IsEscaped": true,
//              "LastObjectSize": "346",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:5:0",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +14,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "e67c93e3e2d4e9d504707bf511bc8bdd6e947b15",
//     +            "Hash": "271d690a0b3349560cbded5a74f61b7040c98731",
//                  "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:16:0"
//              }
//          }
