# gnokey

`gnokey` is a tool for managing https://gno.land accounts and interact with instances.

## Install `gnokey`

    $> git clone git@github.com:gnolang/gno.git
    $> cd ./gno
    $> make install.gnokey

Also, see the [quickstart guide](../../../docs/users/interact-with-gnokey.md).

## Manual Entropy Generation

For maximum security, you can provide your own entropy instead of relying on
computer-generated randomness. Manual entropy generation creates a solemn ritual
that emphasizes the importance of randomness in key generation. This method
ensures your private key's randomness comes from physical sources rather than
computer algorithms. Your input is SHA-256 hashed to create the seed and the
same entropy always produces the same mnemonic.

```bash
# Interactive entropy input
gnokey add mykey --entropy

# Masked input (hides characters as you type)
gnokey add mykey --entropy --masked
```

### Instructions

Generate true random entropy using ONE of these methods:

• **Dice**: Roll a D20 (20-sided die) exactly 42 times
  Example: `18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16 3 8 12 19 2 7 14 5 11 18 1 20 9 4 15 13 17 6 10 16 3 11`

• **Cards**: Shuffle a standard 52-card deck 20 times, then record the full deck order
  Example: `AS 2H 7C KD 3S 9H QC 4D JH 10S 5C 8H AC 2D 7S KH 3C 9D QS 4H JS 10C 5D 8S AH 2C 7D KC 3H 9S QD 4C JC 10H 5S 8D AD 2S 7H KS 3D 9C QH 4S JD 10D 5H 8C 6S 6H 6D 6C`
