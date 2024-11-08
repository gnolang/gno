const { MerkleTree } = require("merkletreejs");
const SHA256 = require("crypto-js/sha256");

var assert = require("assert");

describe("Generate Tree", function () {
  const tests = [
    {
      size: 1,
      expected:
        "cf9f824bce7f5bc63d557b23591f58577f53fe29f974a615bdddbd0140f912f4",
    },
    {
      size: 3,
      expected:
        "1a4a5f0fa267244bf9f74a63fdf2a87eed5e97e4bd104a9e94728c8fb5442177",
    },
    {
      size: 10,
      expected:
        "cd8a40502b0b92bf58e7432a5abb2d8b60121cf2b7966d6ebaf103f907a1bc21",
    },
    {
      size: 1000,
      expected:
        "fa533d2efdf12be26bc410dfa42936ac63361324e35e9b1ff54d422a1dd2388b",
    },
  ];

  for (let tt of tests) {
    describe(`size_${tt.size}`, function () {
      it(`should return ${tt.expected}`, function () {
        const leaves = Array.from({ length: tt.size }, (_, i) =>
          SHA256(`node_${i}`),
        );

        const tree = new MerkleTree(leaves, SHA256, {});
        const leaf = SHA256(leaves[0]);

        const proof = tree.getProof(leaf);
        const root = tree.getRoot().toString("hex");

        assert.equal(root, tt.expected);
      });
    });
  }
});
