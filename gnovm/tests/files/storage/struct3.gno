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
// gno.land/r/storage: 154

// Realm:
// finalizerealm["gno.land/r/storage"]
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:14](6)=
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
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:14",
//              "LastObjectSize": "240",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "RefCount": "1"
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:16](6)=
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
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:16",
//              "LastObjectSize": "241",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:15",
//              "RefCount": "1"
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:6](135)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:6",
//              "LastObjectSize": "254",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,15 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage.Foo"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:15"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:15](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:15",
//     +        "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     -        "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:5",
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
//     -            "Hash": "bb4bcc7028604a7fbdc85dbe1e017703f6c8a528",
//     +            "Hash": "985baca75e3f9c0249bcb63a65cfa9b0a3c8ac03",
//                  "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:16"
//              }
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:5](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:5",
//              "LastObjectSize": "389",
//     -        "ModTime": "13",
//     +        "ModTime": "16",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "342613533eaa1f2d8652ef611d53dc5d8d1801bc",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:15"
//                  },
//                  "Index": "0",
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "344",
//     -        "ModTime": "13",
//     +        "ModTime": "16",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "8c36c4bf5daa89e705234abe4880cab6d2eaf051",
//     +            "Hash": "fc5508b081e5f1ffcd17dd6a84f41f093ba1efdc",
//                  "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:14"
//              }
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1b944a9891f4d61b58cd8cdd043967193ba5441c",
//     +                "Hash": "3119d37cbd490c23a32d259496578ad6900ed36c",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3ae514288350fe29be2d7fe9212fb9d5798ff136",
//     +                "Hash": "00ecaae00214166798c72976c2e24bb7170366b9",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:5"
//                  }
//              },
//     @@ -54,7 +54,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1e1c91dededdda71471dfd8176c4aa9aaca5242a",
//     +                "Hash": "b02f225e5d54582b47d4b84090ca96e11192532b",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:6"
//                  }
//              },
