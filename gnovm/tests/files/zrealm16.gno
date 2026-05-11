// PKGPATH: gno.land/r/test
package test

var (
	a  = &A{A: "here"}
	a2 = &A{A: "here"}
	b  [2]*B
)

type A struct {
	A string
}
type B struct {
	A *A
	B string
}

func init() {
	c := B{
		A: a,
		B: "c",
	}
	b[0] = &c

	d := B{
		A: a,
		B: "d",
	}
	b[1] = &d
}

func main(cur realm,) {
	b[0].A = a2
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:15](5)=
//     @@ -12,7 +12,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -32,7 +32,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15",
//              "LastObjectSize": "441",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     -        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//     -        "RefCount": "1"
//     +        "ModTime": "17",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "15",
//     -        "RefCount": "3"
//     +        "ModTime": "17",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:6](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "LastObjectSize": "378",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "01ee6dd26ce66b6fe21404d36fd173959b219ec6",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Index": "0",
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:14](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//              "LastObjectSize": "336",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "207a677e4346938f113dca8a661ac4325555a3dc",
//     +            "Hash": "3da5946b73b428aecf1ba32383b7dd0f4df567eb",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:10](0)=
//     @@ -13,7 +13,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "4afb55912424a9f5702bc8846c4c4b764559c3f4",
//     +                    "Hash": "1f104dcf7699596665cc35b19219b72d558fe59f",
//                          "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14"
//                      },
//                      "Index": "0",
//     @@ -43,7 +43,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "583",
//     -        "ModTime": "13",
//     +        "ModTime": "17",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:9](0)=
//     @@ -21,7 +21,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "e8634db6099ab7cf368a2954d722bfb505425d60",
//     +            "Hash": "5ae867b75f8c02332e43301b797a40432fb8fb9f",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "34d6abf8dfbd39643b6b7302f42db2c2b4e3559d",
//     +                "Hash": "eecaee5b8f0a613c3163f5bf4b64763594108e8f",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                  }
//              },
//     @@ -54,7 +54,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c96123940e2e20739b7586cde3d099a1f5e83ba9",
//     +                "Hash": "6d245fce22000ca30d9bf8f221c43eac6a254804",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  }
//              },
