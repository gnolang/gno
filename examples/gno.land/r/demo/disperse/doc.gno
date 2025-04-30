// Package disperse provides methods to disperse coins or GRC20 tokens among multiple addresses.
//
// The disperse package is an implementation of an existing service that allows users to send coins or GRC20 tokens to multiple addresses
// on the Ethereum blockchain.
//
// Usage:
// To use disperse, you can either use `DisperseUgnot` to send coins or `DisperseGRC20` to send GRC20 tokens to multiple addresses.
//
// Example:
// Dispersing 200 coins to two addresses:
// - DisperseUgnotString("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150,50")
// Dispersing 200 worth of a GRC20 token "TEST" to two addresses:
// - DisperseGRC20String("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150TEST,50TEST")
//
// Reference:
// - [the original dispere app](https://disperse.app/)
// - [the original disperse app on etherscan](https://etherscan.io/address/0xd152f549545093347a162dce210e7293f1452150#code)
// - [the gno disperse web app](https://gno-disperse.netlify.app/)
package disperse // import "gno.land/r/demo/disperse"
