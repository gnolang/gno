// PKGPATH: gno.land/r/storage_test
package storage_test

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
	crossing()
	SetName("aa")
	record3 = record2
	println(GetName())

}

// Output:
// aa:aa:aa

// Storage:
// gno.land/r/storage_test: 145

// Realm:
// switchrealm["gno.land/r/storage_test"]
// u[334570344df8e77213d2886cb407c37de6da36c1:14](3)=
//     @@ -7,13 +7,13 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "a"
//     +                "value": "aa"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:14",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//              "RefCount": "1"
//          }
// u[334570344df8e77213d2886cb407c37de6da36c1:16](3)=
//     @@ -7,13 +7,13 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "a"
//     +                "value": "aa"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:16",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:15",
//              "RefCount": "1"
//          }
// u[334570344df8e77213d2886cb407c37de6da36c1:6](135)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:6",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//          },
//     @@ -12,6 +12,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage_test.Foo"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Escaped": true,
//     +                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:15"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[334570344df8e77213d2886cb407c37de6da36c1:15](4)=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:15",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "16",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:5",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
//     @@ -12,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "b8f0b5d9326c6936ca77cb517072d622a7c5e083",
//     +            "Hash": "5e9ba4c9bd7488a8918a1b9a0f3e0bf5a7303ffd",
//                  "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:16"
//              }
//          }
