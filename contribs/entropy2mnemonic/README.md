# entropy2mnemonic

Standalone version of `gnokey add --entropy` for generating BIP39 mnemonics from custom entropy. Useful for creating mnemonics for hardware wallets like Ledger or any situation where you want deterministic key generation from your own entropy source.

Uses the same entropy-to-mnemonic conversion as gnokey: SHA-256 hash of your input entropy is used as the seed for BIP39 mnemonic generation.

## Example

```bash
$ entropy2mnemonic
=== ENTROPY TO MNEMONIC CONVERTER ===

This tool generates a BIP39 mnemonic from your custom entropy.
The same entropy will always produce the same mnemonic.

REQUIREMENTS:
- Minimum 27 characters for 160-bit security
- Recommended 43+ characters for better security

GOOD ENTROPY SOURCES:
- Dice rolls: 38+ d20 rolls (e.g., 18 7 3 12 5 19 8 2 14 11...)
- Coin flips: 160+ flips (e.g., HTTHHTTHHHTTHHTHTTHHTHHT...)
- Playing cards: 31+ draws (e.g., 7H 2C KS 9D 4H JS QC 3S...)
- Random typing, environmental noise, etc.

Enter your entropy (press Enter when done):
dice: 18 7 3 12 5 19 8 2 14 11 20 1 9 15 4 13 6 17 10 16 4 8 12 3 7 19 2 11 15 18 5 9 14 6 1 20 13 10 17 4 8 16

Entropy received:
  Length: 122 characters
  SHA-256: 95fa677df9707da96cce5f1b80482a369e96de0e94b9c77dd1e31df82fba6469

Generated mnemonic (24 words):
nominee spring term very amazing start rebel slogan breeze across appear hospital emotion rabbit snack please loop real inmate pet unusual any journey avocado

IMPORTANT: Store this mnemonic securely. It cannot be recovered!
```