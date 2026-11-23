// PKGPATH: gno.land/r/storage
package storage

// sort.Slice moves stored elements rather than copying them, so the realm
// updates only the array and creates or deletes no object.

import "sort"

type item struct {
	Key  int
	Name string
}

var items = []item{{3, "c"}, {1, "a"}, {2, "b"}}

func main() {
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	println(items[0].Name, items[1].Name, items[2].Name)
}

// Output:
// a b c

// Storage:
// gno.land/r/storage: 15

// Realm:
// finalizerealm["gno.land/r/storage"]
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:4](5)=
//     @@ -1,17 +1,6 @@
//      {
//          "Data": null,
//          "List": [
//     -        {
//     -            "T": {
//     -                "@type": "/gno.RefType",
//     -                "ID": "gno.land/r/storage.item"
//     -            },
//     -            "V": {
//     -                "@type": "/gno.RefValue",
//     -                "Hash": "44f80e56c9e3cff5a52d42b266dc50cb27b6958c",
//     -                "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:5"
//     -            }
//     -        },
//              {
//                  "T": {
//                      "@type": "/gno.RefType",
//     @@ -33,12 +22,23 @@
//                      "Hash": "9e7b023fd0ef8ef022812f1ac96bb724fd366921",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7"
//                  }
//     +        },
//     +        {
//     +            "T": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/storage.item"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "44f80e56c9e3cff5a52d42b266dc50cb27b6958c",
//     +                "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:5"
//     +            }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:4",
//              "LastObjectSize": "658",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "RefCount": "1"
//          }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "385",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "17d6e55bc509934576c02b8b200cc7bd4501c68d",
//     +                "Hash": "0fdbf093607eb0cc4c7f8ca7355790cea1a6db19",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:4"
//                  },
//                  "Length": "3",
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "IsEscaped": true,
//              "LastObjectSize": "590",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "7af568e4ac2258dfdf2045d3b216958779c70334",
//     +                "Hash": "d00095ff6034dc840fbea70ff37aa1049e87ceb7",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
