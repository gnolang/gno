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
// u[334570344df8e77213d2886cb407c37de6da36c1:5](3)=
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
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:5",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//          }
// u[334570344df8e77213d2886cb407c37de6da36c1:7](3)=
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
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:7",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:6",
//              "RefCount": "1"
//          }
// u[334570344df8e77213d2886cb407c37de6da36c1:2](135)=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "IsEscaped": true,
//     -        "ModTime": "4",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -25,7 +25,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "6ab157b02caeaa4894b430dfb81ae5d5018190ee",
//     +                "Hash": "3dee67a43ba2708daafa1814a656445fca96fcaf",
//                      "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:5"
//                  }
//              },
//     @@ -41,7 +41,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "d258d9941d042ad81c2e1c0f73a64a224c03afe5",
//     +                    "Escaped": true,
//                          "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:6"
//                      },
//                      "Index": "0",
//     @@ -55,6 +55,16 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/storage_test.Foo"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Escaped": true,
//     +                    "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:6"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              },
//              {
// u[334570344df8e77213d2886cb407c37de6da36c1:6](4)=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:6",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "7",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
//     @@ -12,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "4e95045d28b8adaeb884032eb8bd5f38626c2251",
//     +            "Hash": "5df2d261e7532616baca0bd91e9f6b80fcbc5ba4",
//                  "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:7"
//              }
//          }
