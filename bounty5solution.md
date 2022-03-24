## Problem Definition

here is the problem definition

https://github.com/gnolang/bounties/issues/15


## Solution Breakdown


>Defend against hacking issues that may arise from hardware wallet providers.

hardware wallet provider may not use a RAND number to generate mnemonic. The attacker could use each number in a pre-existing sequence and a counter stored on devices to generate mnemonics that look random but is not. For example, use pre-existing sequence in Pi,  Prime number, or even block hash in the major blockchain.

> Defend against potential weaknesses in the Secp256k1 algorithm. For example, Satoshi's "2^256 - 2^32 - 2^9 - 2^8 - 2^7 - 2^6 - 2^4 - 1" constant is potentially flawed.

we can use the ed25519 algorithm to create a backup key and store it in a separate key base.

> Defend against potential weaknesses in bip32's HMAC-SHA512 function. SHA256 is somewhat economically tested by the Bitcoin hashing algorithm and mining incentives, so the goal here is to provide an alternative that relies primarily on sha256.


> Defend against potential weaknesses in bip39's PBKDF2 function (which also relies on HMAC-SHA512). For example, the PBKDF2 function may have a limited range of outputs, which limits the private keyspace.

We can use a real hkdf with an extract-then-expand scheme.
https://en.wikipedia.org/wiki/HKDF

It is standardized in RFC5869 by Internet Engineering Task Force.

https://datatracker.ietf.org/doc/html/rfc5869

The primitive is implemented in golang.
https://pkg.go.dev/golang.org/x/crypto/hkdf


>Continue to allow the usage of hardware signing devices, and bip32/39 and secp256k1 algorithms for day-to-day usage.

Create a 2/2 multi-sig combined from primary key and backup key
Use a single command to sign transactions with a 2/2 multi-sig.  No need to sign a transaction twice and no need to combine two signatures to gather.


## Implementation explained.

This bounty#5  implementation uses the following priorities to resolve conflicts

Security > Useability > Simple implementation

I extended gnokey to gnokeybk which is backward compatible

#### bkkey sub command:

it generates a backup key:

- Create a backup key using a mnemonic generate on an air gap computer
- Use ed25519 and HKDF to generate a backup key
- Store the backup key in a separate key base file. It provides additional security.  We can even move the backup key store from the air-gap computer to a USB stick after we complete the signing task.

- The backup key info is multi-sig info. It contains ed25519 key and multisig pubkeys that combine primary pubkey and backup pubkey.  Since the attacking point for multi-sig is at the time of combining two keys, we need to make sure only the person holding the primary key can create this backup key info. This is very IMPORTANT.

- This implementation introduces a primary key signature in backup key info.   It uses the primary key to sign the backup info including name, ed25519 privkey armor, and combined multisig pubkeys.  The signature and the primary pubkey are used to prove that the backup info is created by the primary key holder. If the backup key store is altered, it will give errors




#### listbk sub command:

It lists the primary key and backup key from two different key stores.

#### sign sub command:
It retrieves the primary key and backup key from Keystore sign the transaction, combine signatures in one transaction with multisig pubkeys. During the process, it also verifies the backup key integrity stored in the backup key store.

#### changes in the forked code base.

To minimize the impact to the code base before the implementation is reviewed. I wrote all the relevant files in gno/pkgs/crypto/keys/backup_keybase.go and gno/cmd/gnokeybk/  
	 once we review it and approve the implementation. these codes can be merged back to the existing framework.

There are two additional minor updates on the forked code base.
Registered infoBk package
github.com/gnolang/gno/pkgs/package.go


Added backup key multisig address in genesis state
github.com/gnolang/gno/cmd/gnoland/main.go


## Discussions.
> I propose that we allow for the registration of an alternative key based on ed25519 as a backup key that is not used but can be used in case of emergencies when issues arise with the default bip32/39/sec256k1 keys.


This part is tricky since on-chain verification needs to decide if it requires verifying two signatures or just one. Maybe we can force users to use the backup key (two signatures)to sign contracts once it is generated and registered on the Chain. To transfer funds, we may not need to be strict and allow the use of either primary key( one signature) or backup key (a multisig with two signatures)

> Tooling should be provided to allow this alternative backup key to be generated on an air-gapped computer, without the aid of a specialized crypto signing device.

Agree, done

> The (primary) mnemonic for the secp256k1 key must be separate from the (secondary) mnemonic for the ed25519 key, because hardware crypto signers ask for the mnemonic to set up the device, and the second mnemonic should only be entered on air-gapped general computers.

Agree, done

> In Gno, instead of relying on the memo field, we can add a "RegisterAccount" sdk.Msg to provide the backup ed25519 pubkey-hash, and furthermore, we can require the user to register their account w/ backup key before actually using their secp256k1 keys. This way, users can just use a hardware wallet to generate a secp256k1 address to receive funds but are incentivized to register a ed25519 pubkey.

Suggest to store backup multisig key address in registeredAccount

> In the case of issues with the bip32/39/secp256k1 system, the gno.land chain can fork and just use the ed25519 key. Users who have not yet registered an ed25519 backup key would either end up losing their tokens, or possibly the tokens would have to be distributed with real-world KYC etc to catch the hacker who may be trying to reclaim them on the gno.land fork -- this assuming that we come up with a reasonable way to protect the privacy of users while also keeping the recovery accountable. Users who register their account with a backup key would not be affected, and if there are no issues with bip32/39/secp256k1, none of this matters.

With existing implementation and force using the backup once it is generated, we do not need to fork the chain even we know the primary keys are compromised.


> 24-words -> standard kdf & hd derivation -> secp256k1 address
> 24-words (different) -> sha256-based-kdf -> ed25519 address

Agree, done.

> And finally, the gnokey command should be updated to make all of this easier. I like the approach taken by #11 and #14; there just needs to be another gnokey subcommand that bundles these two together (and explains everything, and requires different mnemonics) and produces an unsigned msg for account registration on gno, as well as MEMO-based airdrop-registration on cosmos.

Agree, users can recover the cosmos key on gnoland first and then generate a backup key. The generated backup key address can be stored in "RegisterAccount" sdk.Msg and broadcast to the chain. This way chain can key a record of who generated backup.

## Assumptions, Limitations, and Further Discussions

When the attack happens we probably do not know in advance. We will only discover it after the primary key is distributed and used widely.


For security reasons, no backup key should be altered or recreated. Since the bad guy can create a backup key as well and submit registration transactions to the chain. There is no way to prevent it on Chain. So we are under the assumption that users create backup keys before bad guys take the action


Due above open issues and assumptions, when a backup key is recognized on the chain, the chain should only accept the tx signed by backup key multisig
