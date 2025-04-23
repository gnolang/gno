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
	crossing()
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
// gno.land/r/storage_test: 1

// Realm:
// finalizerealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:13](215)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:13",
//         "ModTime": "0",
//         "ObjectSize": "215",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//         "RefCount": "1"
//     }
// }
// u[334570344df8e77213d2886cb407c37de6da36c1:10](4)=
//     @@ -7,15 +7,15 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "a"
//     +                "value": "aa"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:10",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "ObjectSize": "240",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//     -        "RefCount": "1"
//     +        "RefCount": "0"
//          }
//      }
// u[334570344df8e77213d2886cb407c37de6da36c1:12](6)=
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
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:12",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "ObjectSize": "241",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:11",
//              "RefCount": "1"
// u[334570344df8e77213d2886cb407c37de6da36c1:11](5)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:11",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "ObjectSize": "345",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:5",
//              "RefCount": "1"
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "7df9bf7a04b361753924e514a9864f381c74c053",
//     +            "Hash": "7f434eaa69cd0b01c8a9fa0e0817a330e338ef42",
//                  "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:12"
//              }
//          }
// u[334570344df8e77213d2886cb407c37de6da36c1:6](135)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:6",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "ObjectSize": "259",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage_test.Foo"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "930e867bb908cf8e3317c946866c85d62cc15f65",
//     +                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:11"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[334570344df8e77213d2886cb407c37de6da36c1:3](3)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "ObjectSize": "346",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "6250cc3319a807eddbd7b694e74fd4c71832611c",
//     -            "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:10"
//     +            "Hash": "bf7dafffd0e4a7e506da9770592c2ba4dee42518",
//     +            "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:13"
//              }
//          }
//      }
// u[334570344df8e77213d2886cb407c37de6da36c1:5](-127)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:5",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "ObjectSize": "391",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//     @@ -13,16 +13,6 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage_test.Foo"
//                  }
//     -        },
//     -        "V": {
//     -            "@type": "/gno.PointerValue",
//     -            "Base": {
//     -                "@type": "/gno.RefValue",
//     -                "Hash": "0dd3adc4c8e151857896a710470c9500917e8a5d",
//     -                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:11"
//     -            },
//     -            "Index": "0",
//     -            "TV": null
//              }
//          }
//      }
// d[334570344df8e77213d2886cb407c37de6da36c1:10](-240)
