# GNO.LAND LAWS

_These Laws are governed by the Constitution. Law Amendments require a
Supermajority Decision by GovDAO. If there are any conflicts between the
Constitution and Laws the Constitution takes precedence._

## State Purge Transaction Procedures

State Purge Transactions must be constructed by a fully deterministic and
accountable procedure made available to anyone to run freely without depending
on any external services or APIs. State Purge Transactions must be signed by
authorized signers determined by a future Constitutional Amendment with
limitations, controls, and a system of accountability so as to prevent the
abuse of the chain.

The procedure for the submission of State Purge Transactions may not
necessarily be part of the usual logic of the node software as it may require
more resources than should be made to all nodes and may involve prioritization
based on off-chain tips such as by flagging which may not require any gas fees.
However the node software must be made such that node operators may optionally
automatically check the integrity of state purge transactions and detect
false-positives (due to non-determinism bugs or due to the wrongful signing of
State Purge Transactions). False Positive Purge Detection Transactions may be
signed and submitted by active validators as well as any GovDAO delegated
authority; but whether such transactions have any immediate effect or whether
node operators must check the integrity of the transaction will be determined
by future Constitutional Amendments. False Positive Detection Transactions that
are shown to be valid must be addressed by GovDAO.

All AI models used for the purpose of automated moderation must be registered
on chain by the hash of its bytes and must be static and not automatically
trained and updated with new transactions such that anyone can check the
integrity of State Purge Transactions easily with the exact AI model registered
on chain at the time of purge. AI models may be replaced with newer models or
be trained with blockchain data and periodically be registered with the chain.
All AI models used for the purpose of automated moderation must be made
available for anyone on demand with any reasonable fees for the transmission of
its bytes.

## Private Key Security Guide

GovDAO must maintain a simple guide for users to harden their security including:

 * Discouragement from entering their private key or mnemonic on any online
   computer or mobile internet-capable device, even if the mobile device is
   marketed as being secure for key storage; with education about the ability
   for state actors to add back-doors to such systems.

 * How to obtain an offline computer from a list of approved devices and an
   audited image for the operating system with images committed on chain by
   their hash with bounties for demonstrating a vulnerability in the offline
   computer signing context.

   * Only devices that don't already have any network (including bluetooth)
     capabilities may be included in the list of approved devices.  Software
     disabling of network or bluetooth capabilities is not sufficient.

 * Encouragement to use a 52 deck of cards or 42 rolls of 20-sided dice or the
   equivalent rolls of 6-sided dice to generate custom entropy; with education
   that hardware wallets may be compromised to generate insecure private keys.

Gno.land nor GovDAO nor any entity receiving funding from Gno.land or GovDAO
may not sell any hardware devices except by approved retailers of the
manufacturer.

## Safety Wrapper Contracts

GovDAO must ensure the timely development of conventions, protocols, and
libraries for realm and library logic to support the freezing of user accounts
or native tokens or application tokens/property in such a way to minimize harm
for unrelated parties.

For the purpose of protecting users from theft or loss resulting from exploits,
hacks, or even user error GovDAO must fund for the development of multiple
competing implementations of wrapper contracts that address one or more of the
following concerns:

 * restriction patterns of inflow and outflow of tokens or property
 * pluggable circuit breakers and administration to control inflow and outflow
 * realm upgrading for one or more patterns
 * organization administration

For example, time-based restriction of outflows (a waiting period) allows for a
circuit breaker to prevent a hacker from running off with stolen funds; and
value-based throttling restriction of outflows can add supplementary protection
if the waiting period fails; and such a Safety Wrapper Contract could allow for
outflow restrictions to overlap with inflow restrictions to save time without
compromising on security.

## Quantum Proof Cryptography Roadmap

The following is the roadmap for quantum proofing Gno.land. It may be modified
with a Law Amendment by GovDAO and also the pre-approval of the Oversight Body.

### User Accounts

   * Within 1 year after launch:
     * Determine a protocol to register each account with two Merkle tree hash
       commits of at least 64 distinct Lamport signature public keys with one
       commit using only the sha256 hash function and the other commit using a
       different hash function family such as sha3 (Quantum Recovery Commit)
       that can be used to securely register/sign-for at least 64 new arbitrary
       (e.g. quantum-proof) public key bytes of any length determined in the
       future (e.g. after "D-day") for each commit.
   * Within 2 years after launch:
     * Deploy the aforementioned Quantum Recovery Commit registration system
       and require its usage for all validators and also new user accounts
       before any new transactions are accepted.
     * Determine node, client, and hardware wallet software changes to prevent
       the re-use of public keys for user transactions without relying on new
       or unproven cryptography. Each account is associated with the next
       deterministic 20-byte or larger address committed by each transaction
       based on the strictly sequential account sequence number.
     * Begin work with reputable cryptographers with recent relevant
       publications to determine the number of bytes (currently 20) needed and
       ideal hash function for account addresses as determined by the best
       theoretical attacks, and assuming the most optimistic/advanced estimates
       and projections for quantum computers developed in secret by nation
       state actors.
   * Within 3 years after launch:
     * Deploy the aforementioned node, client, and hardware wallet software
       changes to prevent the re-use of public keys for user transactions.
   * Within 4 years after launch:
     * Disable the re-use of public keys for all user transactions.
     * Freeze all accounts that do not yet have a Quantum Recovery Commit. This
       is to prevent future mass hacking of accounts by quantum computers which
       can destroy the economy of the chain.
   * Users may continue to use secp256k1 hash addresses as long as those
     addresses are not re-used, the address has sufficient byte length given
     the best known theoretical attacks, and attacks for a given address has
     not yet been provably demonstrated within the unbonding period of
     Atom.One. This is to prevent requiring users to adopt new (e.g. new
     quantum-proof) cryptography algorithms that are yet unproven or may prove
     to be insecure in the future.
   * New quantum-proof public key systems must be strictly additive to the
     usual public key system until proven by industry for at least 5 years.

### Validators

Validator signing and verification have different requirements than user
accounts (e.g. they can assume larger memory and compute capacity than hardware
wallets; and they must remain secure even after public key exposure for
light-client/IBC-bridges and for double-signing slashing/jailing mechanics) so
require a different path for post-quantum readiness.

   * Within 2 years after launch:
     * Investigate changes to Tendermint2 with NT,LLC to support pluggable
       quantum proof cryptography.
   * Within 3 years after launch:
     * Complete or approve an initial draft modification of Tendermint2 also
       approved by NT,LLC to support pluggable quantum proof cryptography.
   * Within 4 years after launch:
     * Support and test final release of Tendermint2 by NT,LLC to support
       pluggable quantum proof cryptography.
   * Within 5 years after launch:
     * Require validators to run Tendermint2 with quantum proof cryptography.
       If Atom.One does not allow post-quantum Tendermint2 by this time and
       Gno.land runs on Atom.One ICS, Gno.land must migrate away from Atom.One
       ICS.

Tendermint2 like the GnoVM is owned by and funded by NT,LLC. The funding for
the development of the above is NT,LLC. Failing the above GovDAO may decide to
fork Tendermint2 on a branch with a unique name with the above as priorities.

## New Cryptographic Primitives or Implementations

Any new cryptographic primitives or implementations must be approved by
a Cryptography Committee approved by GovDAO with at least 3 T1 or T2 members
who are respected and have well-cited publications in the field of cryptography
and at least a doctorate degree in cryptography or relevant fields in
mathematics, and at least 1 T1 or T2 members who are equally expert in the
field of quantum computing.

Such new cryptographic primitives or implementations have a need that is not
met by the existing core software.

New cryptographic primitives for when an old primitive is deemed
insufficiently secure (e.g. when quantum computers break secp256k1 given
pubkey) then both the old and new pubkeys must be independently derived and
both required for a testing period of 7 years with significant economic
incentives such as with bounties rewarded for cracking challenge keys. This
helps prevent the adoption of primitives or implementations with unknown or
secret exploits from posing systemic risk.

## Formal Verification of Go/Gno

Once the market cap of $GNOT exceeds $8B GovDAO must pass a Law Amendment to
include provisions for the formal verification of Go or Gno programs with the
pre-approval of a reputable researcher with publications about formal
verification using open source tools already used by industry for formal
verification of programs.

If meaningful progress of such cannot be made during any 2 years of funding
another qualified researcher must be consulted for development with a new team.

## Development of Open Hardware

Once the market cap of $GNOT exceeds $5B GovDAO must pass a Law Amendment to
include provisions for the development of 100% open hardware (and 100% open
firmware/software) devices for the following categories in order:

 * A device that sits between an online computer or mobile network-capable
   device and hardware signer such that all communication between the
   aforementioned devices can be inspected (the Hardware Packet Inspector).
   This reduces the chances that a backdoor in the hardware wallet may result
   in theft or loss. If no suitable 100% open hardware CPU chip is available a
   RISC-V architecture based chip may be used instead until one is available.

 * Hardware signer that is separate from but compatible the Hardware Packet
   Inspector. This may use a single chip that is proprietary to hold the
   private key and logic for unlocking based on pin, but it must be a module
   contained within completely open hardware and open firmware/source such that
   a vulnerability or backdoor of the single chip can be limited.

 * Validator hardware signer to preserve the private key of the validator and
   prevent double-signing.

 * CPU chip based on the RISC-V architecture (or related reputable open source
   forks) for use in any of the above devices.

 * A device for the hardening of chips such that keys cannot be extracted
   easily from these chips by forensic analysis.

The open hardware and open software/firmware must be made available primarily
by a copyleft license compatible with the latest Gno Network GPL license or a
Gnu copyleft license with the appropriate attribution terms for hardware such
that the product is free of patent restrictions and all modifications to the
funded source must also be made available with no proprietary modifications
allowed except for personal use, and any derivative works and any manufacture
must give prominent credit to the original designer(s) of the hardware.

The funding may go toward investment in an existing reputable hardware
developer or contract with one but the resulting work must be made available
such that after its release anyone may build upon the product with appropriate
attribution.

All open hardware must be formally verified by standard formal verification
software before release.

No one is required to use the these products.

## Forensic Analysis of Common Hardware

Furthermore once the market cap of $GNOT exceeds $5B GovDAO must pass a Law
Amendment to include provisions for funding an in-house team to accountably
verify the integrity of commonly used hardware products of Gno.land users,
especially the forensic analysis of the chips of those hardware products.

 * The devices must be acquired externally by a variety of means so as to
   prevent the prediction by manufacturers or retailers. If a device cannot be
   acquired by the user from a physical store without identification, such a
   device cannot be trusted at all. For example, certain AT&T devices can only
   be purchased by AT&T through the mail, and they are tampered with in transit
   when being delivered to targetted individuals. These devices will be
   published in a blacklist curated by GovDAO or a delegate commitee. These
   devices will not be tested.

 * Users may also provide hardware to be inspected by this team for pay.

 * The analysis must be done with a video and photographic recording so as to
   ensure that the analysis is correct. The video and photographs must be of
   sufficient quality so as to be correlated with each other and the
   photographs of sufficient quality that they can be inspected by the public.

 * The shipment of devices to the forensic analysis team must be done in such a
   way that the origin of the devices can be verified during the recording and
   photography.

All of the procedures, software, and materials used by the forensic analysis
team must be made public and free such that any team may use them to offer
their own forensic analysis.
