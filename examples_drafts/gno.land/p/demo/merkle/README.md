# p/demo/merkle

This package implement a merkle tree that is complient with [merkletreejs](https://github.com/merkletreejs/merkletreejs)

## [merkletreejs](https://github.com/merkletreejs/merkletreejs)

```javascript
const { MerkleTree } = require("merkletreejs");
const SHA256 = require("crypto-js/sha256");

let leaves = [];
for (let i = 0; i < 10; i++) {
  leaves.push(SHA256(`node_${i}`));
}

const tree = new MerkleTree(leaves, SHA256);
const root = tree.getRoot().toString("hex");

console.log(root); // cd8a40502b0b92bf58e7432a5abb2d8b60121cf2b7966d6ebaf103f907a1bc21
```
