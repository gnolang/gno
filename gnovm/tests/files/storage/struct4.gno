// PKGPATH: gno.land/r/storage_test
package storage_test

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
// &(struct{("bb" string)} gno.land/r/storage_test.Foo)

// Storage:
// gno.land/r/storage_test: -20

// Realm:
// switchrealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:8](214)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:8",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//         "RefCount": "1"
//     }
// }
// u[334570344df8e77213d2886cb407c37de6da36c1:5](1)=
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
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:5",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//     -        "RefCount": "1"
//     +        "RefCount": "0"
//          }
//      }
// u[334570344df8e77213d2886cb407c37de6da36c1:7](3)=
//     @@ -7,13 +7,13 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "b"
//     +                "value": "bb"
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
// u[334570344df8e77213d2886cb407c37de6da36c1:6](2)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:6",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "134e725d6cffbcdc489a987817be1d59af9ca569",
//     +            "Hash": "2935d4e5350c371a2d9ec161eb48da70cb6d657a",
//                  "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:7"
//              }
//          }
// u[334570344df8e77213d2886cb407c37de6da36c1:2](0)=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "IsEscaped": true,
//     -        "ModTime": "4",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -25,8 +25,17 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "6ab157b02caeaa4894b430dfb81ae5d5018190ee",
//     -                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:5"
//     +                "Hash": "e9a89d647b4af39916fc3cf8a5d8fc3ede124b09",
//     +                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:8"
//     +            }
//     +        },
//     +        {
//     +            "T": {
//     +                "@type": "/gno.PointerType",
//     +                "Elt": {
//     +                    "@type": "/gno.RefType",
//     +                    "ID": "gno.land/r/storage_test.Foo"
//     +                }
//                  }
//              },
//              {
//     @@ -41,20 +50,11 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "c60ddc71d28cb7757a230038cf10f977d76291b3",
//     +                    "Hash": "f86d78491e03b34570bb589e4d255724cf6e4683",
//                          "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:6"
//                      },
//                      "Index": "0",
//                      "TV": null
//     -            }
//     -        },
//     -        {
//     -            "T": {
//     -                "@type": "/gno.PointerType",
//     -                "Elt": {
//     -                    "@type": "/gno.RefType",
//     -                    "ID": "gno.land/r/storage_test.Foo"
//     -                }
//                  }
//              },
//              {
// d[334570344df8e77213d2886cb407c37de6da36c1:5](-240)
