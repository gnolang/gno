// PKGPATH: gno.land/r/test
package test

import "math"

type Account struct {
	Balance uint256
	Debt    int256
}

var (
	supply   uint256
	delta    int256
	zero     uint256
	acct     Account
	accounts = map[string]*Account{}
	history  [3]int256
	amounts  []uint256
	ptr      *uint256
)

func init() {
	supply = math.MaxUint256
	delta = math.MinInt256
	acct = Account{Balance: 1 << 200, Debt: -1}
	accounts["alice"] = &Account{Balance: 7, Debt: -7}
	history = [3]int256{-1, 0, 1}
	amounts = []uint256{1, 1 << 128}
	v := uint256(42)
	ptr = &v
}

func main(cur realm) {
	// The values written by init were persisted and are read back here.
	println(supply, delta, zero)
	println(acct.Balance, acct.Debt)
	println(accounts["alice"].Balance, accounts["alice"].Debt)
	println(history[0], history[1], history[2], amounts[0], amounts[1], *ptr)

	supply--
	delta++
	acct.Debt *= 2
	accounts["alice"].Balance <<= 250
	history[1] = math.MaxInt256
	amounts = append(amounts, 3)
	*ptr = 0
}

// Output:
// 115792089237316195423570985008687907853269984665640564039457584007913129639935 -57896044618658097711785492504343953926634992332820282019728792003956564819968 0
// 1606938044258990275541962092341162602522202993782792835301376 -1
// 7 -7
// -1 0 1 1 340282366920938463463374607431768211456 42

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:23](394)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "4194304"
//             },
//             "V": {
//                 "@type": "/gno.Bits256Value",
//                 "value": "AQ=="
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "4194304"
//             },
//             "V": {
//                 "@type": "/gno.Bits256Value",
//                 "value": "AQAAAAAAAAAAAAAAAAAAAAA="
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "4194304"
//             },
//             "V": {
//                 "@type": "/gno.Bits256Value",
//                 "value": "Aw=="
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:23",
//         "LastObjectSize": "394",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "281",
//     -        "ModTime": "16",
//     +        "ModTime": "22",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.Bits256Value",
//     -            "value": "//////////////////////////////////////////8="
//     +            "value": "//////////////////////////////////////////4="
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "281",
//     -        "ModTime": "16",
//     +        "ModTime": "22",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.Bits256Value",
//     -            "value": "gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
//     +            "value": "gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAE="
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:17](5)=
//     @@ -17,14 +17,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.Bits256Value",
//     -                "value": "//////////////////////////////////////////8="
//     +                "value": "//////////////////////////////////////////4="
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17",
//              "LastObjectSize": "367",
//     -        "ModTime": "0",
//     +        "ModTime": "22",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:19](36)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.Bits256Value",
//     -                "value": "Bw=="
//     +                "value": "HAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
//                  }
//              },
//              {
//     @@ -24,7 +24,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19",
//              "LastObjectSize": "343",
//     -        "ModTime": "0",
//     +        "ModTime": "22",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:20](62)=
//     @@ -15,6 +15,10 @@
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "2097152"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.Bits256Value",
//     +                "value": "f/////////////////////////////////////////8="
//                  }
//              },
//              {
//     @@ -31,7 +35,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:20",
//              "LastObjectSize": "383",
//     -        "ModTime": "0",
//     +        "ModTime": "22",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:12](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "LastObjectSize": "378",
//     -        "ModTime": "16",
//     +        "ModTime": "22",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,11 +19,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ea9148f92d88a2537f08690cae7570bd01cdeed4",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:21"
//     +                "Hash": "234c66d567042712e84cdf04908305a15270f94c",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:23"
//                  },
//     -            "Length": "2",
//     -            "Maxcap": "2",
//     +            "Length": "3",
//     +            "Maxcap": "3",
//                  "Offset": "0"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:22](-21)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:22",
//              "LastObjectSize": "247",
//     -        "ModTime": "0",
//     +        "ModTime": "22",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//              "RefCount": "1"
//          },
//     @@ -10,10 +10,6 @@
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "4194304"
//     -        },
//     -        "V": {
//     -            "@type": "/gno.Bits256Value",
//     -            "value": "Kg=="
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1747",
//     -        "ModTime": "22",
//     +        "ModTime": "23",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "864f038ca46f242a8d7c7e0d46adf022de9ff5c1",
//     +                "Hash": "b86b8a33395c16b9e514bbd24e644843be5091a5",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "bd138e86132d1314554c0f61449b02d82f60e2bc",
//     +                "Hash": "3f70233ba11f23b4eacb1e9e48e486a877e1ec60",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
//     @@ -76,7 +76,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "0678365b929d3d47c6af135faa143dd0d1ceee09",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                  }
//              },
//     @@ -95,7 +94,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d0e12a85f18816c689734c34bf1099fdfc72d3e2",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//                  }
//              },
//     @@ -105,7 +103,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "4c443624c0092340fd2eaa4e14e9fac84da30dc5",
//     +                "Hash": "54db4a3869acb82162d85b9c0d496c0db64d3dd0",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//                  }
//              },
//     @@ -115,7 +113,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "cc3006108508bab25162c38ad0eb96762f832a4b",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//                  }
//              },
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:6](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "LastObjectSize": "345",
//     -        "ModTime": "16",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "4dc905a8d7f61d2052f9b2c07a02fa286d890fe9",
//     +            "Hash": "0558d699ac63ea6f8be42e844002d3d8765e01ee",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:17"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:18](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18",
//              "LastObjectSize": "341",
//     -        "ModTime": "0",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "0f0f49c280796062f79ac80f9486aa88499dfd2b",
//     +            "Hash": "d0862fef9de0c9331edf576a9f327de7facb7da9",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:19"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:9](0)=
//     @@ -24,7 +24,7 @@
//                              "@type": "/gno.PointerValue",
//                              "Base": {
//                                  "@type": "/gno.RefValue",
//     -                            "Hash": "e5c87749a03f0e72a8b253bba0e731663638d3f4",
//     +                            "Hash": "ccfa23566502c8e2d9cf8bc3304b68031cbd63f9",
//                                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:18"
//                              },
//                              "Index": "0",
//     @@ -37,7 +37,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "LastObjectSize": "458",
//     -        "ModTime": "16",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:8](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "410",
//     -        "ModTime": "22",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -23,7 +23,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "de5d25df1270c03684f015efa0c3ad92ad7fb0df",
//     +            "Hash": "ff57bb20dfc1cb6eac3196bdcd48d4ebd24e5835",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:10](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "354",
//     -        "ModTime": "16",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "eccd52323472651e14c2544b1464e1d6b7487ac0",
//     +            "Hash": "2ae3ba567251d0714788093af1926cc29cedbce5",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:20"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:13](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//              "LastObjectSize": "377",
//     -        "ModTime": "16",
//     +        "ModTime": "23",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "bdf8254f2bf5e0ea0853e3c297ba82e61971ab5c",
//     +                "Hash": "8fad259209007fe9a6dcc3cb20edb97c43375c67",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:22"
//                  },
//                  "Index": "0",
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:21](-327)
