# GNO.LAND CONSTITUTION

_This Constitution is still a draft and until $GNOT is transferrable any
Constitutional Amendments shall require the pre-approval of NT,LLC, or an
amendment shall be approved with a Majority Decision by GovDAO if proposed by
NT,LLC. When there are any ambiguities or conflicts within this document NT,LLC
reserves the right to clarify them until $GNOT is transferrable. After $GNOT is
made transferrable the normal mechanism of Constitutional Amendments will
apply._

## Declaration of Intent

We the gnomes of Gno.land, to bring to life a virtual world of structured
information and logic expressed in a mutiuser language based system, for
bringing light to darkness by hosting all that will be known, hereby establish
this Constitution for Gno.land.

## Terms

 * AIB,Inc: All in Bits, Inc.

 * NT,LLC: NewTendermint, LLC.

 * Governing Documents: This Gno.land Constitution, its Laws, and any
   applicable Mandates and Bylaws (altogether Governing Documents).

 * Core Software: The minimal set of reference or standard software needed for
   the Gno.land blockchain and its users. See "Software Security".

 * Constitution: This Gno.land constitution document including modifications by
   Constitutional Amendments. This Constitution is considered independent of
   the Atom.One Constitution although some portions of the Atom.One
   Constitution may be referenced here for clarity or context.

 * Constitutional Amendment: A modification to this living Constitution by a
   Constitutional Majority Decision of GovDAO and any other requirements as
   defined in this Constitution; composed of diffs.

 * Constitution Only: When something is governed by the "Constitution Only" its
   governing rules must be specified within the living Constitution directly;
   Not in any Laws, Bylaws, or Mandates.

 * Majority, Supermajority, and Constitutional Majority Decision: See "GovDAO
   Decisions" and "Common DAO Spec" // XXX also link to committees

 * Oversight Body: Initially NT,LLC as represented by a multisig account, but
   eventually represented by an Oversight DAO on Gno.land.

 * Treasury: See "Gno.land Treasuries".

 * GovDAO DAO: "All GovDAO DAOs" or "GovDAO DAO Set" refers to GovDAO and all
   descendant subDAOs ultimately created or managed by GovDAO. "A GovDAO DAO"
   (or "relevant GovDAO DAOs") refers to one DAO within the GovDAO DAO Set (or
   any number of DAOs within the GovDAO DAO Set) that has a Mandate that
   matches the given scope.
 
 * Core DAO: GovDAO, the Oversight Body DAO, or any subDAO created or managed
   by any Core DAO implemented on Gno.land conforming to the Common DAO Spec.

 * ICS: Inter-Chain-Security, or the general service of validating a chain with
   the validator set of another chain, or a subset thereof.

 * Simple Replicated ICS: ICS but where the same validator set of the original (hub) 
   is used to secure the other (consumer) chain.

 * SDDCAs (Storage Deposit Discoungt Credit Accounts): Excess $GNOT deposits
   from the redution in price of $GNOT/byte deposit ($GNOT Storage Deposit
   Price). See "Excess $GNOT Deposits".

 * Open Soruce IP: Intellectual Property that is made available freely under
   the Gno Network GPL or any liberal license compatible with the Gno Network
   GPL.

 * Fully Audited: See "Software Security".

 * XXX

## Genesis Allocation

At Gno.land Genesis there will be one billion $GNOT tokens.

 * Airdrop1:               35.0% - from partial Cosmos governance snapshot 3 years ago
 * Airdrop2:               23.1% - from recent AtomOne snapshot prior to launch
 * EcosystemContributors:  11.9% - for prior and future Gno.land ecosystem contributors
 * Investors:               7.0% - reserved for current and future investments
 * NT,LLC:                 23.0% - of which a significant portion is allocated for prior loans

$GNOT will not be transferrable initially except for whitelisted addresses.
Whitelisted addresses include "GovDAO" and "Investors" funds and any additional
addresses dedicated for faucets.

The 7% (qualfied, private) investors allocation will be held by NT,LLC in a
segregated account. Proceeds of sales of these tokens will go toward NT,LLC for
past or future development of Gno.land, Gno, Tendermin2, other Core Software,
and ecosystem development.

GovDAO is responsible for distributing the $GNOT of the Ecosystem Contributors
Treasury allocation to prior and future Gno.land ecosystem contributors (as
well as those contributing to the blockchain stack, including external
contributors to Tendermint2, GnoVM, Gno.land server and tooling, GnoWeb) with
the exclusion of existing GovDAO members. At least one-quarter and up to
one-third of the GovDAO 11.9% genesis allocation will be distributed to prior
contributors by GovDAO Supermajority Decision. Present GovDAO members are not
eligible for any allocation from the EcosystemContributors genesis allocation.

NT,LLC owes around $20M worth of $GNOT to All in Bits,Inc at the "fair market
value" of $GNOT at a 25% discount for a "fair market value" which is yet to be
determined as of the date of this writing.

## Gno.land Tokenomics

> $ATONE  : like a Bitcoin miner\
> $PHOTON : like a BTC token\
> $GNOT   : storage deposit token

First a primer on Atom.One tokenomics. $ATONE and $PHOTON are tightly coupled:

 * $ATONE is an inflationary staking token capped at 20% inflation.
 * $PHOTON is a deflationary transaction gas-fee (spam-prevention) token.
 * If all $ATONE burned there would be one billion $PHOTON.
 * Burning is one way; $PHOTON cannot be burned back to $ATONE.
 * $ATONE targets 2/3 to be staked, thus inflationary staking "rewards" are not
   income.
   * Any tax authority that says otherwise is robbing you.

While $PHOTON has no voting rights and cannot be used for staking on Atom.One,
it nevertheless has unique advantages enforced by the Atom.One constitution:

 * All transaction fee payments on Atom.One must be paid with $PHOTON.
 * Atom.One ICS Consumer Chains (aka ICS shards)
   including Gno.land after migration must accept $PHOTON as the gas-fee token.
   * Each may also accept other gas-fee tokens but ultimately the chain must
     pay $PHOTON to Atom.One for security.
 * Atom.One ICS sovereign (consumer) chains may exit (change validator-sets)
   Atom.One and become self-validating or become hosted by another ICS hub.

Gno.land will launch independently of Atom.One so Gno.land will initialy
require $GNOT for transaction fee payment. Before and after Gno.land migrates
to be hosted by Atom.One ICS $GNOT will function as a byte storage deposit
token.

 * Every transaction that ends up increasing the amount of persistent state
   will require a bond deposit of $GNOT.
 * Every transaction that ends up freeing up persistent state space will
   receive a refund of $GNOT.
 * One billion $GNOT corresponds to 10TB of persistent state space.
 * The $GNOT inflation schedule will never change, thus the total created $GNOT
   will never exceed 1.333~ billion $GNOT.
 * The $GNOT Storage Deposit Price (per byte) will never increase.
 * The $GNOT Storage Deposit Price will be such that the total remaining $GNOT.
   and all future inflationary $GNOT never exceeds the size of a typical
   consumer internal hard-disk or solid-state drive. See also "Gno.land Storage
   Capacity".

Gno.land is obligated to migrate to be hosted/secured by Atom.One ICS when it
is deemed ready according to this Constitution by GovDAO by Supermajority
Decision.

After migration to Atom.One ICS hosting Gno.land should pay the Atom.One chain
in $PHOTONs underneath the hood as the Atom.One constitution requires; an AMM
exchange module should exist on the Gno.land shard/instance to facilitate the
internal exchange of collected $GNOT to $PHOTON needed to pay Atom.One for its
ICS services.

Once Gno.land migrates over to Atom.One after the Gno.land <> Atom.One IBC
connection is complete and Atom.One ICS MVP is implemented $ATONE will be the
staking and governance token on Atom.One (but with no voting rights for
Gno.land itself) and no voting rights for Gno.land itself, and $PHOTON will be
primarily the CPU-time gas token paid to Atom.One (which in turn pays for all
that is necessary to secure the chain via ICS validation), and $GNOT the
dedicated byte-storage deposit token on Gno.land. Gno.land will become a key
ICS consumer chain on Atom.One especially in the beginning even as Atom.One is
free to offer its ICS services to other applications unrelated to Gno.land and
$GNOT, or even those forked of Gno.land and the GnoVM in the future. 

## $GNOT (Deflationary) Inflation

From the date of launch every year 33.33*(0.9^Y) million $GNOT 3.333% of the
Gno.land Genesis $GNOT supply will be inflated continuously as follows (where Y
is the year from launch starting from 0). Any inflationary $GNOT not
transferred will accrue until the Gno.land software is updated to transfer such
funds.

 * After   3 years this represents  90.32 million $GNOT.
 * After  10 years this represents 217.09 million $GNOT.
 * After 100 years this represents 333.29 million $GNOT.

At most there will be 1.333... billion $GNOT, representing a total inflation of
one third of the genesis $GNOT distribution. This amount cannot change even
with amendments to the Gno.land Constitution. This makes $GNOT a deflationary
token similar to Bitcoin.

The inflated tokens will be distributed as follows:

 * One third to the Core Software Treasury.
 * One third to the Essential Services Treasury.
 * One third to the Ecosystem Contributors Treasury.

## Gno.land Treasuries

The Gno.land Treasuries are as follows:

 * Core Software Treasury
 * Essential Services Treasury
 * Ecosystem Contributors Treasury
 * Validator Services Treasury (ValTreasury for short)
 * GovDAO Pay Treasury (PayTreasury for short)
 * Ecosystem Audit Treasury
 * Recompense Treasury
 * Reserve Treasury

A Treasury is defined as account or DAO that can hold funds where its type,
structure, governance (if any), and purpose are specified in the Constitution
Only. If a Gno.land Treasury is defined without governance or conditions its
funds can only be used with a Constitutional Amendment. _The Reserve Treasury
in particular is reserved for future uses._

A Treasury with a given purpose cannot be used for any other purpose even if
there are excess funds in the Treasury account (or DAO). If GovDAO passes a
Constitutional Amendment that violates this princple (such as by the deletion
of this term) it is the responsibility of the Oversight Body to reject such
proposals. Constitutional Amendments for clarification of language in the
spirit of the terms prior to any such amendments are allowed for cases where
the original language is ambiguous. _The exception to this rule is as per the
disclaimer at the top of this draft Constitution--the set of treasuries and
their purposes, or even this term, may change under certain conditions until
$GNOT is made transferrable._

No spending proposal should be voted on unless an accountability framework is
first defined by a Constitutional Amendment pre-approved by the Oversight Body
such that all spending is reviewed by an independent committees or entities
before and after the spending decision.

No subDAO of a top-level Core DAO may make funding decisions from the Core
Software Treasury directly unless otherwise specified in the Constitution.

Funds from Gno.and Treasuries may be transferred to another Core subDAO
ultimately managed by GovDAO such that the funds of any Core subDAO may be
frozen or returned to the origin Treasury at any time with a Majority Decision
of the responsible Core DAO.

Funds that are transferred or assigned to any of these treasuries but not yet
used (e.g. by a decision to fund an entity or a DAO) may not be clawed back or
transferred to any other DAO or treasury nor burned without a Constitutional
Amendment.

All spending for software or other IP must be for Open Source IP.

### Core Software Treasury

These funds are reserved for the development of Core Software.

All funding decisions from the Core Software Treasury require a Supermajority
Decision by GovDAO. 

The Core Software is the minimal set of reference or standard software needed
for the Gno.land blockchain and its users and includes (and must prioritize):

 * Gno.land node
 * GnoWeb+Alts
 * Tendermint2
 * GnoVM
 * Go
 * Atom.One IBC and ICS
 * Official standard reference browser extension wallet
 * Official standard hardware wallet software
 * All other software mentioned in the Constitution Alone

All Core Software must be released under the Gno Network GPL License with the
appropriate strong attribution clauses as determined by the owner of the
software, except for pre-existing software that is already released with a
compatible open source license.

Funding for Tendermint2, the GnoVM, and the gno cli is primarily the
responsibility of NT,LLC. Funding for Atom.One IBC and ICS is primarily the
responsibility of the Atom.One chain. That said, GovDAO should contribute to
these projects (and any other projects of the Core Software owned by external
entities) within reason as needed to complete the objectives mentioned in the
relevant Governing Documents; and thereafter for as long as the software is
deemed Core Software GovDAO must ensure the continuous and timely monitoring of
future improvements and updates by reasonable funding if necessary from the
Core Software Treasury to keep Gno.land and its users safe.

Funding decisions for any implementations of any consensus engine (such as
Tendermint2), GnoVM, the "gno" cli, or any alternative implementations of Gno
(whether compiler or interpreter) require NT,LLC pre-approval for a period of 4
years. Funding for the monitoring of improvements and updates to these software
do not require NT,LLC pre-approval.

GovDAO is required to use a legal entity such as NT,LLC or any new entities
(Proxy Entity) to keep ownership of IP on behalf of the Gno.land chain. New
Proxy Entities must be proven to be subservient to GovDAO in accordance with
the Constitution such that no transfers of IP nor any changes to license terms
may occur without the prior approval of the relevant GovDAO DAO and Oversight
Body. Each IP repo must have a top-level Markdown file describing the proxy
relationship (except for Tendermint2 and the GnoVM) 

Refactoring of projects out of the genesis monorepo should be coordinated with
NT,LLC but in such a way that preserves the history of all prior commits
relevant to all the files of each project (via the git "--follow" flag) or
otherwise as agreed with NT,LLC. The ownership of the genesis monorepo and its
containing organization (e.g. "gnolang" on Github) and repos under this
organization is by NT,LLC; changes to the ownership of any of these repos
require NT,LLC approval. That said Gno.land is not required to use these
particular repos as long as the commit history is preserved in any forks
thereof. The expectation is for NT,LLC and GovDAO to come to agreement about
NT,LLC serving as the Proxy Entity for some or most of the repos already hosted
under the organization.

### Essential Services Treasury

These funds are reserved for the development of Essential Services.

All funding decisions from the Core Software Treasury require a Supermajority
Decision by GovDAO. 

Essential Services are the set of services rendered by machine or man that are
deemed to essential for the operation of Gno.land and for users to securely
interact with Gno.land and includes (and must prioritize):

 * Blockchain explorers
 * Blockchain indexers
 * User (customer) support
 * Official community activities 

Funding from the Essential Services Treasury for any software services may only
go toward Open Source IP that is fully audited. They may not be used for any
development or auditing unless the development or auditing is necessary and
cannot be funded by other means.

Funding from the Essential Services Treasury may also go toward validation
services on an emergency basis if the Validator Services Treasury is
insufficient.

### Ecosystem Contributors Treasury

These funds are reserved for rewarding the creators of External Contributions
defined as contributions of Open Source IP mentioned in any of the Governing
Documents created by individuals or entities who are not owners of the original
IP nor employees of the owners nor stakeholders of the original IP; or any Gno
code uploaded to Gno.land; and in all cases were not otherwise contracted to
develop said IP and at the time of contribution was not a member of GovDAO.

All funding decisions from the Ecosystem Contributors Treasury require a
Supermajority Decision by GovDAO. 

All funding from the Ecosystem Contributors Treasury must go to past
contributions. That is, no funding may go toward future contributions.

All funding decisions from the Ecosystem Contributors Treasury must be based on
well defined heuristics as defined in the Constitution Alone, or based on
quality as assessed by the review of council members of the relevant GovDAO
DAO. Naturally council members must abstain from voting on matters with any
conflict of interests, but the Oversight Body is specifically responsible for
(and has the authority for) ensuring that council members do not collude to
reward each other unfairly in any given proposal nor over multiple proposals.

No heuristic for funding shall be based on code line-count nor the number of
individual contributors.

No funding shall be made in proportion or algorithmic relation to runtime
metrics of gas, storage, nor fees; nor to any measure of financial value (e.g.
managed by a realm or in relation to any transaction); nor to metrics of usage
by import; except that such metrics and measures may be used for the purpose of
categorization or filtering.

All funding from the Ecosystem Contributors Treasury must go to contributors
whose real human identity is known and recorded in accordance with the
Constitution or relevant Laws of Gno.land.

### Validator Services Treasury

The Validator Services Treasury may only be used to pay for validation
services.

No Core DAO shall be responsible for voting for funding decisions from the
Validator Services Treasury except immediately after genesis until automated
payment mechanisms are implemented.

See also "Gno.land Pre-Migration Validators" and "Essential Services Treasury".

### GovDAO Pay Treasury

These funds are reserved for paying GovDAO members who are actively
contributing to Gno.land. Participation in the governance of GovDAO itself is
not considered active contribution.

All funding decisions from the GovDAO Pay Treasury require a Supermajority
Decision by GovDAO (naturally in accordance with the Constitution).

No payment may go to GovDAO members for any contributions older than 3 months.
That is, retroactive compensation is prohibited.

T1 and T2 members who are actively contributing full-time must get paid
equally. T3 members must get paid strictly less by comparison.

T3 members may not get paid unless there exists in the GovDAO Pay Treasury
sufficient funds to pay all T1 and T2 members for 7 years after launch taking
into account the latest estimated projections including new T1 and T2 members.
That said, T3 members may get paid by other means outside of the GovDAO Pay
Treasury.

Payment for any member from the GovDAO Pay Treasury may not exceed the 90th
percentile of senior software architect roles in the second highest paid city
globally.

Members already employed by another company will not receive any compensation
unless they disclose their compensation to GovDAO T1 members via a process
defined in the Governing Documents; afterwards they may be compensated up to
50% of the usual limit to top up their net payment to the usual limit.

If there are not enough funds in the GovDAO Pay Treasury to pay all GovDAO
members for the next quarter, payment must be reduced equally by up to 10% of
the usual amount for the next quarter. If this reduction is not sufficient then
all T3 members must lose their funding (if any). If this reduction is still not
sufficient then T2 members must lose their funding for the next quarter based
on seniority. If this reduction is still not sufficient then T1 members must
lose their funding for the next quarter based on seniority.

Members who lose their pay due to inadequate funds of the GovDAO Pay Treasury
do not automatically lose their membership; and must not be required to work
more than 25% of full-time to maintain any status regarding activity.

### Ecosystem Audit Treasury

The Ecosystem Audit Treasury may only be used to fund of the auditing of code
deployed to Gno.land by Qualified Auditors, or to assess and reward Bonded
Auditors who submit valid Bonded Vulnerability Reports.

All funding decisions from the Ecosystem Contributors Treasury require a
Supermajority Decision by GovDAO.

GovDAO by Supermajority Decision may choose to burn tokens from the Ecosystem
Audit Treasury at a rate not exceeding 10% a year.

### Recompense Treasury

The Recompense Treasury may only be used to recompense victims of exploits and
fraud.

All funding decisions from the Recompense Treasury require a Supermajority
Decision by GovDAO (naturally with members with any conflict of interests
abstaining).

There is no obligation to compensate anyone; and furthermore no Core DAO may
make guarantees of recompensation to anyone.

Recompensation decisions must be preceded by a thorough analysis of the problem
and tasking of a task-force to recover any ill-gotten gains and at least two
weeks for all relevant parties to review the analysis and task-force for
approval.

GovDAO by Supermajority Decision may choose to burn tokens from the Recompense
Treasury at a rate not exceeding 10% a year.

### Reserve Treasury

No funding decisions are allowed from the Reserve Treasury without a GovDAO
Constitutional Majority Decision.

## Gno.land Revenue

Transaction fees collected in $GNOT or $PHOTON on Gno.land are called Revenue.
The Revenue is distributed according to the following rules in order:

 1. ValTreasury is funded with 75% of Revenue if ValTreasury has less than 1 year of runway, or
 2. ValTreasury is funded with 50% of Revenue if ValTreasury has less than 4 years of runway, or
 3. ValTreasury is funded with 25% of Revenue if ValTreasury has less than 7 years of runway, or
 4. ValTreasury is funded with 10% of Revenue if ValTreasury has more than 7 years of runway.

The portion of Revenue that is not allocated to ValTreasury is called Remaining Revenue 1.

 5. PayTreasury is funded with 75% of Remaining Revenue 1 if PayTreasury has less than 1 year of runway, or
 6. PayTreasury is funded with 50% of Remaining Revenue 1 if PayTreasury has less than 4 years of runway, or
 7. PayTreasury is funded with 25% of Remaining Revenue 1 if PayTreasury has less than 7 years of runway, or
 8. PayTreasury is funded with 10% of Remaining Revenue 1 if PayTreasury has more than 7 years of runway.

PayTreasury has strictly lower priority than ValTreasury because GovDAO members
can theoretically still function without pay, whereas the blockchain cannot
function securely without validators. The PayTreasury runway must take into
account future projected growth of the paid T1 and T2 members.

The portion of Remaining Revenue 1 that is not allocated to PayTreasury is
called Remaining Revenue 2.

 9. Essential Services Treasury is funded with min(Remaining Revenue 2, 25% of
    Revenue).
 10. Reserve Treasury is funded with the remainder from above.

Essential Services Treasury has strictly lower priority than ValTreasury or
PayTreasury because there is some elasticity to the variety of services that
can be offered, much of which should become mature and finalized; and some of
the essential services could be migrated to be offered by all of the validators
instead (especially with the help of GovDAO members); and services should
generally pay for themselves with micropayments or subscriptions or by other
means.

## $GNOT Burn Mechanisms

The $GNOT token should not be used as a storage deposit token or for any other
chain except the original Gno.land chain, nor be used as a staking/bonding
token as this limits the utility (accessible storage capacity) of Gno.land
itself.

$GNOT is only considered "burned" and according to approved burn mechanisms or
transactions (like $ATONE "burn" to $PHOTON except $GNOT burning does not yield
any new tokens or rights on Gno.land). The automatic burning of $GNOT in a
every Realms' SDDCAs (Storage Deposit Discoungt Credit Accounts) are the only
automatic burn mechanism of $GNOT but the Gno.land Constitution may also
approve of other official burn mechanisms that are reasonable, VOLUNTARY, and
EXPLICIT. Such burning mechanisms may be used to yield storage deposit tokens
on other chains.

XXX Include a standard burn address and/or burn transaction type.

The $GNOT token should not be used as a storage deposit token or for any other
chain except the original Gno.land chain, nor be used as a staking/bonding
token as this limits the utility (accessible storage capacity) of Gno.land
itself. It may make sense to burn $GNOT via approved burn mechisms to acquire
secondary tokens that can be used for storage deposit on other GnoVM (and
non-Gno) chains hosted on Atom.One. See "$GNOT Burn Mechanisms".

## Excess $GNOT Deposits

When the $GNOT storage deposit rate per byte is decreased this results in
excess $GNOT deposits per realm. This is not considered part of Revenue.

When the $GNOT storage deposit rate decreases (not automatically by the $GNOT
burn mechanism described below, but by the decision of GovDAO to lower the
rate) 25% of the excess $GNOT goes into the Ecosystem Audit Treasury, 25% of
the excess $GNOT goes into the Recompense Treasury, and the remainer goes to
the each seggregated virtual account per realm ($GNOT of the realm's SDDCA) to
be used only for future discounts on transactions for that specific realm.
These SDDCA $GNOT tokens may be transferred to other SDDCAs by rules defined in
the Governing Documents but their $GNOT may never be withdrawn even upon
freeing storage, and transfers must be initiated by the present authority of
the realm or containing organization. 

DDCA $GNOTs may be burned automatically by a rate set by an amendment to the
Gno.land Constitution not to exceed 10% a year. This is to prevent stagnant
$GNOT from limiting the allocated storage capacity of Gno.land and thereby
reducing its utility. $GNOT burned in this way will also reduce the $GNOT
storage deposit rate automatically but not trigger any of the the mechanisms
described here.

See also "Ecosystem Audit Treasury" and "Recompense Treasury".

## Gno.land Pre-Migration Validators 

Until Atom.One ICS is ready Gno.land the validator selection mechanism is
determined by GovDAO Supermajority Decision.

Priority should be given to Atom.One validators and Gno.land core developers.

The number of validators prior to migration shall not exceed 50.

No validators may operate on any cloud hosting provider, but must run their own
hardware where they have physical access to the server at all times (such as in
a co-location provider with 24/hour access).

Atom.One migration is not contingent on all of its validators running their
their own hardware as above, but GovDAO may impose a requirement by Majority
Decision for Atom.One to have a completed roadmap specified to get there.

Atom.One ICS shall not be deemed suitable unless the Gno.land chain
remains whole (not part of any "mega-block" system where the consensus engine
process is shared with other applications) and Gno.land may migrate away from
Atom.One by on-chain transactions.

If Atom.One validators do not largely (> 90% by voting power) run on their own
hardware where validators have physical access to their server at all times 2
years after migration or 4 years after after Gno.land launch whichever is sooner,
GovDAO with a Supermajority Decision may choose to fund the development of a
fork Atom.One in the likeness of the Cosmos -> Atom.One fork with a relevant
governance proposal on Atom.One (with preference to the class of voters who
voted in favor of this requirement); provided that GovDAO first submits
reasonable proposals to Atom.One that get rejected by Atom.One.

Likewise if the Atom.One staking incentive system is not such that the top
validator has at most three times the voting power of the lowest validator 2
years after migration or 4 years after Gno.land launch whichever is sooner,
GovDAO with a Supermajority Decision may choose to fund the development of a
fork Atom.One in the likeness of the Cosmos -> Atom.One fork with a relevant
governance proposal on Atom.One (with preference to the class of voters who
voted in favor of this requirement); provided that GovDAO first submits
reasonabe proposals to Atom.One that get rejected by Atom.One.

In the case of an Atom.One fork initiated by GovDAO by the above, the fork need
not run any chain except those that run Gno. The constitution of the fork shall
be determined by Supermajority Decision of GovDAO and respect the rights of
NT,LLC with respect to the "Gno" word mark.

Pre-migration and post-migration validators are paid from the Validator
Services Treasury. See also "Validator Services Treasury".

## Atom.One ICS Migration

Gno.land is obligated to migrate to Atom.One when declared ready by GovDAO
simple majority decision.

In the case of a fork of Atom.One the target chain may be changed by a
Constitutional Amendment to be one of the forks of Atom.One. Any GovDAO members
who are involved in the creation or development of such a fork or have any
additional relative ownership of staking tokens or fee tokens of the fork are
considered to have a conflict of interest and must disclose so and abstain from
voting. If a quorum cannot be reached due the quorum requirement may be waived
by the Oversight Body.

## GovDAO Structure

There are three tiers: T1, T2, and T3.

T1 is the highest tier, T3 the lowest.

 * T1: self-selecting "core" with supermajority vote from T1.
 * T2: selected by GovDAO w/ T3 abstaining, with simple majority vote.
 * T3: permissionless invitation from T1 and T2 according to a delegation
   mechanism.

The maximum age for any member is 70; after is automatically withdrawn.

T1, T2, and T3 membership can only be withdrawn by a GovDAO Supermajority
Decision WITH CAUSE and Oversight Body pre-approval. Such members are
considered "ejected". Ejected members are no longer eligible for membership in
GovDAO.

Members may resign at any time with a signed transaction. Resigned members may
be reinstated with a Supermajority Devision of GovDAO.

T1 members may not be actively contributing to the project but they must still
be active in voting or their membership shall be suspended if they miss 3
GovDAO proposal votes in a row until they return to activity with a simple
declaration. The activity status of T1 or their suspension shall not be deemed
cause for removal from T1. In this way T1 is like "tenure", and generally the
T1 members are expected to have made significant contributions to the project.
The exception to this rule is described in "US Preference Option".

## GovDAO Voting Power.

In general,
 * a T1 member gets 3 votes.
 * a T2 member gets 2 votes.
 * a T3 member gets 1 vote.

However,
 * T2 class is capped at 2/3 the voting power of T1 class.
 * T3 class is capped at 1/3 the voting power of T1 class.
 * --> T1 1/2, T2 1/3, T3 1/6 (unless...)

Example 1:
 * T1 100 members --> 300 VP, 3 votes per member
 * T2 100 members --> 200 VP, 2 votes per member
 * T3 100 members --> 100 VP, 1 votes per member

Example 2:
 * T1 100 members --> 300 VP, 3 votes per member
 * T2  50 members --> 100 VP, 2 votes per member *
 * T3  10 members -->  10 VP, 1 votes per member *

Example 3:
 * T1 100 members --> 300 VP, 3 votes per member
 * T2 200 members --> 200 VP, 1 votes per member *
 * T3 100 members --> 100 VP, 1 votes per member

Example 4:
 * T1 100 members  --> 300 VP,   3 votes per member
 * T2 200 members  --> 200 VP,   1 votes per member *
 * T3 1000 members --> 100 VP, 0.1 votes per member *

### T1 Membership Size

The target T1 membership size is a minimum 70 members in 7 years.

If the target minimum is not reached, at most 2 members may be elected by the
usual process of election every quarter.

If the target minimum is not reached AND there are Qualified Candidates AND two
quarters have passed with no new T1 members elected, 1 additional Qualified
Candidate may be elected by GovDAO with T1 abstaining in the following quarter.

If the target minimum is not reached AND there are Qualified Candidates AND
four quarters have passed with no new T1 members elected, 1 additional
Qualified Candidate may be elected by GnotDAO.

None of the above per-quarter appointment count limitations apply for the US
Preference Option.

### T2 Membership Size

The T2 membership size target minimum is floor(size(T1)/4). The T2 membership
size target maximum is 2 x size(T1).

While size(T2) can be greater than 2 x size(T1) or less than 2 x size(T1), no
more members may be added to T2 if size(T2) is greater than 2 x size(T1).  This
limitation does not apply for the US Preference Option.

### T3 Membership Size

T3 membership is decided automatically based on a delegation system where:

 * T1 members get 3 invitation points.
 * T2 members get 2 invitation points.
 * T3 members get 1 invitation points.

At least 2 invitation points from at least 2 members must be delegated for T3
membership. Invitation points are whole numbers; they are not divisible.

The T3 membership is determined automatically based on the current GovDAO
membership and the latest delegations. Delegations may be changed at any time;
thus a T3 member may lose their membership immediatey as a consequence of
undelegation, and this may cause another T3 member to lose their membership
concurrently.

### Payment to GovDAO Members

See "GovDAO Pay Treasury".

## GovDAO

GovDAO is the primary governing body of Gno.land with limitations by the
Oversight Body and other terms and conditions of this Constitution.

### GovDAO Decisions

A majority decision of GovDAO requires more than 1/2 of voting power of T1, T2,
and T3 members according to the scoring mechanism defined in this Constitution.

A supermajority decision of GovDAO requires more than 2/3 of voting power of
T1, T2, and T3 members according to the scoring mechanism defined in this
Constitution.

A constitutional majority decision of GovDAO requires more than 9/10 of voting
power of T1, T2, and T3 members according to the scoring mechanism defined in
this Constitution.

Some decisions may require additional conditions to be satisifed:

 * When a pre-approval is required the proposal must attach a valid
   cryptographically signed signature of pre-approval of the exact proposal
   body bytes. If no such automated mechanism is implemented on chain the
   proposal must be rejected by all members unless the aforementioned signature
   exists and is independently verified by each voting member. For example, a
   Constitutional Amendment requires the pre-approval of the Oversight Body.

 * When a post-approval is required the proposal must be appropriately
   constructed with a field set that prevents any automated action from taking
   place upon the passage of the proposal until the relevant post-approval
   conditions are met. If no such automated mechanism is implemented on chain
   the post-approval condition must be treated as a pre-approval condition. If
   all relevant conditions are not satisifed the proposal must be rejected by
   all members.

### GovDAO Membership Criteria

#### General Membership Criteria

These GovDAO Membership General Criteria apply for all GovDAO member
candidates:

 * No candidate is eligible who is or had been associated with any entity or
   organization that has confidential membership and requires members to
   protect each other (e.g. Freemasonry). If an existing member joins such they
   must declare a "spiritual conflict of interest" and resign immediately.

 * No candidate is eligible who has or has had any association with any
   government intelligence agency or program, any defense contractor, nor any
   law enforcement agency. Candidates may have served in a military but they
   may not be actively serving.

 * No candidate is eligible who has or has had any affiliation with a communist
   or fascist or Nazi or Neo Nazi nor Zionist party or organization; nor any
   gang.

 * If an existing member gains any association that violates the above general
   requirements they must cease all GovDAO related activity and immediately
   declare their association within 24 hours and resign.

 * No candidate is eligible who has been convicted of a violent, property,
   white-collar, organized, or cyber crime.

   * Any existing member(s) who gets convicted of these crimes must declare so
     and resign or provide the details of the charges and the judgement.
     Thereafter GovDAO minus these members must vote on whether to eject the
     member with a majority decision.

The general requirements for T1, T2, and T3 membership are as follows:

 * All members must be publicly identifiable persons and declare their country
   of residence and citizenshipa and provide timely updates when it changes.

 * T1 members must meet T1,T2,T3 criteria.

 * T2 members must meet T2,T3 criteria.

 * T3 members must meet T3 criteria.

 * T1 criteria includes expertise in categories, significant contributions, and
   demonstration of value alignment, and when active must make public
   appearances promoting Gno.land or publications related to their
   contributions to Gno.land.
   
 * T2 criteria includes expertise in categories and continuing contributions
   incluing public appearances promoting Gno.land or publications related to
   their contributions to Gno.land.

 * T3 criteria includes significant experience in categories and continuing
   contributions.

 * T1 and T2 members are added via individual proposals, each with a markdown
   resume/portfolio application document.

 * Any members who have been proven to have lied or not disclosed material
   information must be demoted or ejected with a majority decision with a
   pre-approval from the Oversight Body contingent on proof and evaluation.

 * Ejected members are never again eligible for membership in T1, T2, or T3.

#### Diversification Requirements

The Diversification Requirements exists to prevent any single country or union
from dominating GovDAO thereby ensuring a globally decentralized governance
body.

 * No new candidates for T1 (or T2) may be proposed if their country or union
   of citizenship represents more than one third of T1 (or T2).

 * No new candidates for T1 (or T2) may be proposed if their country or union
   of residence represents more than one third of T1 (or T2).

#### US Preference Option

The US Preference Option exists as an exception to the Diversification
Requirements and any other restrictions on membership size limitations or rate
of election, but must satisfy the General Membership Criteria as any other
candidate.

 * NT,LLC has the option to appoint as many US sole-citizen T1 (or T2) members
   until there are as many US citizen members of T1 (or T2) of any other
   country or union after such appointment.

 * Dual citizens shall be counted as fractions where the denominator is the
   number of citizenships and the numerator is one. Automatic eligibility for
   citizenship based on birthrights shall count as citizenship.

 * Appointment may happen at any time and is not limited by the Diversification
   Requirements. Specifically the appointment may be made concurrently with any
   new appointments to T1 (or T2) or ejections from T1 (or T2) as long as the
   appointment is announced prior to the conclusion of the election or ejection
   process.

 * NT,LLC may decide at its option to delegate the US Preference Option to the
   subset of current T1 and T2 members who are US citizen residents (US
   Preference Option Default Committee). These members shall have voting power
   ratio of 3:2. The delegation is temporary (per quarter) unless otherwise
   specified. The US Preference Option Default Committee may only elect up to 2
   T1 members and 2 T2 members per quarter.

 * If NT,LLC fails to exercise the US Preference Option when it is otherwise
   possible (e.g. there are eligible Qualified Candidates who are US citizen
   residents) for a continuous period of 2 years AND there are at least 5 US
   citizen residents in T1 or T2, the US Preference Option defaults to the US
   Preference Option Default Commitee.

 * The US Preference Option may not be changed even by a Constitutional
   Amendment except with pre-approval by NT,LLC. This ensures that the will of
   the founder to ensure US representation in GovDAO is preserved at all times.

#### Conflict of Interest Policy

All Core DAO members must annually sign and circulate to all T1 members a
conflict of interest disclosure document.

All Core DAO members who have any conflict of interests with any proposals must
declare their conflict of interest and abstain from voting on the spending
proposal.

See also "General Membership Criteria" for additional terms related to conflict
of interests.

Violations of the Conflict of Interest Policy as determined by the Oversight
Body must result in punitive measures as determined by the Governing Documents
including ejection, suspension, or demotion of membership, or the clawback of
funds. That said, any punative measure by the Oversight Body may be vetoed or
overruled by a GovDAO Supermajority Decision at any time.

## Oversight Body

The Oversight Body is responsible for ensuring that all proposals (except
proposals to replace Oversight Body DAO members) adhere to the Governing
Documents with priority given to the Consitutiton in case of any conflicts.

The Oversight Body is initially NT,LLC as represented by a multisig account
approved by NT,LLC. NT,LLC multisig signers must have an agreement with NT,LLC
and only sign transactions explicitly pre-approved by NT,LLC. In the case of a
breach of this requirement its signature shall have no effect or its effect
must be reversed promptly by GovDAO before any other decision.

Within 2 years after launch NT,LLC must delegate its Oversight Body role to a
DAO (the Oversight Body DAO) on Gno.land, or by by declaring its members in
accordance with all requirements.

If after 2 years after launch NT,LLC does not determine the Oversight Body DAO
or its members, GovDAO shall remind NT,LLC to determine the Oversight Body DAO
members. If after a month such members are not decided by NT,LLC, the Oversight
Body DAO shall be determined by a Supermajority Decision of GovDAO, and no
further decisions may be made by GovDAO until an Oversight Body DAO is
determined that satisfies the requirements.

The Oversight Body DAO committee Requirements are as follows:

 * One Oversight Body DAO member shall be from T1 or T2 (First Oversight Member).
 * One Oversight Body DAO member shall be from T1 or T2 (Second Oversight Member).
 * One Oversight Body DAO member shall not be a GovDAO member but otherwise be
   a Qualified Candidate with no significant conflict of interests such as by
   being invested in a competing crypto project (Third Oversight Member).

Furhermore,

 * Each member gets 1 vote. All decisions require 2 YES votes.
 * Two or more members shall not reside in the same country at the time of
   appointment.
 * Two or more members shall not be citizens of (or be automatically eligible
   for citizenship by birthright of) the same country or union.
 * All Oversight Body DAO members must fully understand the Constitution and
   Laws as assessed in a live interview test with T1 members.
 * Each member shall serve a maximum term of three years, thereafter shall not
   be eligible for re-election for another three years.
 * All members shall be considered working full time and paid as a typical T1
   GovDAO member (including the Third Oversight Member) from the PayTreasury.

After the initial establishment of an Oversight Body DAO:

 * The First Oversight Member may be replaced at any time by a Supermajority
   Decision of GovDAO.
 * The Second Oversight Member may be replaced at any time by NT,LLC.
 * The Third Oversight Member who has served already for three months may be
   replaced at any time by a Supermajority Decision of GnotDAO (not GovDAO).
 * The Third Oversight Member who has served already for three months may also
   be replaced at any time by a Supermajority Decision of GovDAO AND the
   pre-approval of NT,LLC.

After any election by GnotDAO the candidate must pass the live interview test
with T1 members (as determined by a Majority Decision by GovDAO), and the test
must be recorded and result shared with the public. The Third Oversight Member
candidate who gets rejected in the live interview cannot be re-elected for
another one year.

Any Oversight Member who gets replaced shall be deemed to have served their
full term and cannot be re-elected for another three years.

A vacancy (e.g. from resignment, incapacitation, disqualification, ejection
from GovDAO etc) must be replaced before any Constitutional Amendment or Law
gets passed; and furthermore must be replaced within 30 days before further
proposals get passed save for any urgent Node Software Upgrades related to
software bugs or Transaction Replay Forks.

The Oversight Body DAO members committee may not self-mutate except when a
vacancy arises after which the two remaining members may self-elect by
consensus a temporary Qualified Candidate (the Temporary Oversight Member)
without restriction. This Temporary Oversight Member is expected to get
superceded (or be voted in) by the usual election rules, but such a member
shall not be deemed to have been "replaced": they may be elected again soon
after. Temporary Oversight Members may vote to elect more Temporary Oversight
Members with any more vacancies.

With the exception of any Oversight Body DAO election proposals, the Oversight
Body or its subDAOs shall have the authority to (and must block) any decision
by GovDAO or Core DAOs if such decisions are determined to be:

 * in violation of the Constitution or the spirit of the Constitution
 * in violation of any Laws
 * in violation of any Bylaws or Mandates of any Core DAOs

The Oversight Body may block any proposals that have otherwise passed in the
prior month unless otherwise specified or a pre-approval was already granted.
This is especially important for proposals that immediately pass due to a
supermajority decision. XXX improve this by improving the Common DAO Spec.

The Oversight Body does NOT have the sole authority to transfer, spend, freeze,
or burn any funds or property.

## GnotDAO

After 3 years after the launch of Gno.land GovDAO must have implemented
or chosen an implementation of GnotDAO where:

 * GnotDAO is writtenin Gno.
 * GnotDAO runs on a non-upgradeable (immutable) or system realm.
 * Free trading $GNOT may be bonded by users.
 * Users may use bonded $GNOT to vote on any number of valid proposals.
 * Proposals are made from a list of templates as defined by this Constitution
   and are pre-approved by at least two existing GovDAO members.
 * Proposals run with the same rules and duration as Atom.One proposals
   including any self-adjusting deposit limits to reduce spam.
 * At the end of all proposal voting periods of proposals voted on by bonded
   users the $GNOT may be returned to the user immediately.

No Constitutional Amendments may inflate the $GNOT supply above what is already
declared even with Constitutional Amendments.

## Role and Rights of NT,LLC

NT,LLC retains the excusive rights to word mark and brand "Gno", and "\*Gno\*".
NT,LLC grants GovDAO the right to use the "Gno.land" brand, but NT,LLC is
responsible for managing all domain and subdomains that include the "Gno" word
mark unless otherwise delegated as decided by NT,LLC. In the case of the
"gno.land" domain and subdomains NT,LLC must ensure that it points to the
Gno.land chain, unless the Gno.land identity ceases to exist as defined in this
Constitution.

For a period of 4 years after launch NT,LLC has the sole authority to determine
reasonable guidelines for the content of the main entry pages ("gno.land" and
"www.gno.land"). Thereafter the content of these main entry pages are
determined by this Constitution and applicable laws and overseen by the
Oversight Body. All Qualified Forks of Gno.land shall be listed on these main
pages at all times.

Gno.land and its Qualified Forks may only use the "Gno" word mark and brand and
the "Gno.land" identity and domain and "gnoland\*" chain ID for as long as Gno
is the only contracting language on the chain as determined solely by NT,LLC;
and for as long as GnoWeb is used to render the contents; or unless allowed by
NT,LLC pre-approval on chain by cryptographic signatures by its multisig. Only
NT,LLC may determine the identity of the Gno language for any version and may
authorize releases of the GnovM and GnoWeb. The implementation of Gno.land may
not require the GnoVM if after GnoVM is finalized the alternative
implementation is identical with the GnoVM as determined by NT,LLC and each
release of the alternative implementation is approved by NT,LLC on chain by
cryptographic signatures by its multisig.

Nothing prohibites Gno.land from changing its identity and adopting a different
name that does not include the "Gno" wordmark and brand. GovDAO may choose to
do so by a Constitutional Amendment. However in this case NT,LLC reserves the
right to determine a fork of the Gno.land chain with any modifications to the
constitution and with a new governance body. Such a fork shall be considered a
Qualified Fork even if its governance member set is completely independent of
the orignal GovDAO members.

No Constitutional Amendment shall be valid that restricts the rights of NT,LLC
or reduces its powers or authority as declared in this Constitution nor alter
the US Preference Option without the express permission of NT,LLC. This
includes any modifications to the structure and voting rules of GovDAO.

All GovDAO members agree to help enforce these rights of NT,LLC.

## Principle of the Chain

In all cases the latest released Gno.land node software shall be able to run
the transactions from the first Genesis launch until current lastest
transaction. This is achieved by the following:

 * A chain upgrade may have a sequential number in the chain ID.
 * Each unique chain ID has its own Genesis State & transactions from index 0.
 * The last transaction for a given chain ID results in the End State.
 * The Genesis State of the next chain ID is composed of {End State, Migration
   Script, Other Genesiss Params}.
 * The Migration Script is applied to the End State to form a part of the next
   Genesis State.
 * A folder with subfolders for each chain ID in sequence where each subfolder
   includes the blockchain of blocks of txs and a partial Genesis State that
   omits the End State of the previous chain ID; and also the last node release
   for that chain ID.
 * A simple bash script that to derive the latest state from the first Genesis
   by running all txs in sequence from the first chain ID to the last.

Blockchain nodes that have the full transaction history will always be able to
sync to the latest state hash from genesis using the latest released node
software and possibly also including previous node software releases. This is
to preserve integrity. If validators and nodes are not allowed to store any
offending transaction data they may prune past transactions as long as they
maintain a recent qualfiied snapshot as determined by this Constitution or
GovDAO and all transactions since the snapshot such that the latest blockchain
state may be derived from the snapshot.

### Gno.land Node Upgrades

All upgrades to the Gno.land Node Software must be for Fully Audited Open
Source IP (as any official software release) approved by GovDAO Supermajority
Decision that includes the exact commit hash of the software release. The only
exception to the Fully Audited requirement shall be for emergency security
patches as approved by Law.

No node operator shall be required to run Gno.land Node Software that cannot be
deterministically built with the source even in the case of security issues.

### Gno.land Storage Capacity

The $GNOT Storage Deposit Price will be such that the total remaining $GNOT.
and all future inflationary $GNOT does not exceed the size of a typical
affordable consumer internal hard-disk or solid-state drive available for
purchase by anyone for a PC; where such a storage drive is manufactured by
independent manufacturers of at least three independent manufacturers and in
three different countries; and such a storage drive is at least as good as the
following in key performance metrics:

 * Western Digital 10TB WD\_Black Performance Internal Hard Drive HDD - 7200
   RPM, SATA 6 Gb/s, 512 MB Cache, 3.5" - WD102FZBX

This is to keep the blockchain state at an accessible level for newcomers,
developers, and hobbiests and also accounts for any future potential economic
collapses.

### State Purge Transactions

State Purge Transactions are blockchain transactions for deleting state from a
Gno.land realm or package. Transactions that depend on state purged by State
Purge Transactions shall fail with a special transaction response code.  The
Merkle-tree root hash shall be derivable as if the data was there even after
purge by State Purge Transactions. This helps preserve the integrity of the
chain state for valid use-cases and makes it easier to undo when needed.

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
node operators must check the integrity of the transaction shall be determined
by future Constitutional Amendments. False Positive Detection Transactions that
are shown to be valid must be addressed by GovDAO.

All AI models used for the purpose of automated moderation shall be registered
on chain by the hash of its bytes and shall be static and not automatically
trained and updated with new transactions such that anyone can check the
integrity of State Purge Transactions easily with the exct AI model registered
on chain at the time of purge. AI models may be replaced with newer models or
be trained with blockchain data and periodically be registered with the chain.
All AI models used for the purpose of automated moderation shall be made
available for anyone on demand with any reasonable fees for the transmission of
its bytes.

### Addressing Core Software Bugs

See "Transaction Replay Forks".

### Addressing Hacks, Theft, and Exploits.

See "Transaction Replay Forks".

## Amendments to the Constitution 

All amendments (modifications) to this Constitution incuding changes to any
definitions shall require a Constitutional Amendment pre-approved by the
Oversight Body and passed by a Constitutional Majority Decision of GovDAO
following all the rules of this Constitution.

Amendments to this Constitution must belong to one of severeal categories:

 1. Rewording of portions of the Constitution for clarity or refinement while
    maintaining the structure of the Constitution with no other additions or
    deletions.
 2. Restructuring to move portions of the Constitution for legibility without
    any changes to wording except for section headings or titles.
 3. Additions and deletions of portions of the Constitution without any other
    restructuring or rewordings.

Furthermore the Constitution shall be one single markdown file and the latest
constitution shall be present in the repository under docs/CONSTITUTION.md.

Each amendment shall be composed of up to three diff patches, one for each of
the three categories in the order as declared above.

## User Rights and Limitations

User rights, protections, and limitations will be determined by future
Constitutional Amendments provided that freedom of speech will be protected
with the following exceptions:

### Banned Offensive Material

Pornographic material is defined as follows and can be further restricted by
Law:

 * photographs or animated or AI rendered depictions of humans or animals or
   humanoid figures engaged in sexual activity or depicted in an arousing way
   including games or movies that include such material; except of classic art
   hand-painted prior to the year 1985 or classic sculptures hand-made prior to
   the year 1985.

 * photographs or animated or AI rendered depictuions of naked minors under the
   age of 18 or humanoid figures that could reasonably construed to be a minor
   including games or movies that include such material; except of classic
   sculptures hand-made prior to the year 1985.

Gore material is defined as follows:

 * realistic depictions of gore whether photographic, video, animated, or AI
   rendered, and whether embedded in a movie or in a game; and shall exclude
   material that is submitted for educational purposes especially those related
   to political conflicts or abuses of power.

Harmful harmful (material harmful for users) is defined as:

 * material that is determined to result in hacks or exploits of user devices
   unbeknownst to the user. This excludes material that is submitted for
   educational purposes and includes relevant disclaimers to make clear that
   the material is harmful and is unlikely to result in accidental hacks or
   exploits of user devices.

Pornographic, Gore, and Harmful material are altogether called Banned Offensive
Material. Banned Offensive Material as determined by the relevant GovDAO DAO or
as determined by the Oversight Body may be filtered by GnoWeb or the blockchain
node in API responses in such a way that anyone who runs GnoWeb or the
blockchain node can also replicate the filter without pay or conditions and
without relying on any external service. This shall not include hatespeech or
misinformation: neither GovDAO nor the Oversight Body nor any Core DAOs may
determine what is hatespeech or misinformation. This is to ensure freedom of
speech.

For the avoidance of doubt, any files released to the public as a matter of
public accountability regarding the Epstein files may reside on gno.land but
with complete redaction of the offending materials (e.g. redacted photos are
OK, explicit photos are not). The files are a matter of national and
international security, and as such they belong on Gno.land as do any
discussions. The privacy of individuals mentioned in those files is not yet
protected by this Constitution but should be clarified with an amendment.

Only Banned Offensive Material and harmful material may be purged
from the blockchain state by marking code packages or state objects (including
entire realms) as such with a [State Purge Transaction](#state-purge-transactions).

The Princple of the Chain shall be preserved at all times; that is, the
blockchain of blocks of transactions will not itself be modified (but the state
may be with subsequent transactions). See also "Principle of the Chain".

Code packages and realms that are primarily designed or used to bypass these
checks may also be frozen with a State Freeze Transaction with a future
Constitutional Amendment that defines the procedure and limitations but in no
case shall any state be purged except for the Banned Offensive Material. A porn
realm may have all of its pornographic material purged but comments about the
porn do not fall in this category so therefore cannot be purged. This is to
prevent abuse and censorship of legitimate discussions.

### Intellectual Property and Privacy

DMCA or GDPR or similar that are concerned with intellectual property or
privacy will not affect the Principle of the Chain; this is because such laws
can be misused to censor information that the public ought to know. As with the
aforementioned other offending material each validator or node is responsible
for deleting old blockchain transaction history as needed. GovDAO may pass a
Constitutional Amendment to include DMCA or GDPR issues in the class of
material that may be pruged from blockchain state with a Purge Transaction but
only after a heirarchical bonded system of manual review first takes place to
filter for a strict subset of DMCA or GDPR or similiar requests in accordance
with terms defined in the Constitution under "User Rights and Limitations". In
no case shall there be an automated system that purges such state, nor any
guarantees of timeliness of processing such material be offered or required.

Notwithstanding the above, such material may be filtered by GnoWeb.

### $GNOT Deposits for Purged State

Any $GNOT byte deposit tokens of state purged from the blockchain state with
State Purge Transactions (after a proving period of 6 months without being
marked on-chain as a false-positive purge by any False Positive Purge Detection
Transaction nor determined to be a false-positive by this Constitution or the
Oversight Body) may only be used for the purpose of developing or acquiring
self-contained open-source AI models and tools for automated or semi-automatic
purging of state; or be used to pay a moderation team as determined by this
Constitution. The confiscation of deposited $GNOT for such purposes shall
require a special Purge Deposit Confiscation Transaction signed by any parties
as determined by this Constitution, and shall fail if the purge occured less
than 6 months ago or if the state was marked since by a False Positive Purge
Detection Transaction.

No more than half of such confiscated $GNOT deposits may be used to fund for
any manual moderation--half or more must be used for the development of
automated systems. The Oversight Body may redirect a third of these funds to
fund for the development or acquisition of alternative moderation systems that
conform to all requirements of this Constitution if it determines that the
existing moderation system is not effective, or has an unacceptable amount of
false-positives; but the adoption of any alternative system shall require a
GovDAO majority decision contingent on there being no more false-positive
purges and no more false-negatives as determined by this Constitution and the
GovDAO voting members. The Oversight Body may also block the submission of any
or all Purge Deposit Confiscation Transactions at any time if it determines
that it is being misued; after which GovDAO is responsible for the continued
funding for moderation from its own treasury until the Oversight Body resumes
the submission Purge Deposit Confiscation Transactions.

User or realm tokens or property that are not $GNOT deposit tokens of purged
state whether native tokens or Gno application derived may not be confiscated
by any means, though they may be frozen by State Freeze Transactions.

### Hacks, Thefts, and Exploits

User or realm tokens or property that are determined to be derived from any
hacks, thefts, or exploits and are not determined to have already been
transferred to unrelated parties may be frozen by a State Freeze Transaction or
the funds used to recompense associated victims as determined by a future
Constitutional Amendment and governed by the Constitution Alone.

Each recompense decision shall be independent of another and require a
Supermajorithy Decision of GovDAO. The public shall be provided sufficient
information to judge the soundness of each decision.

### Transaction Replay Forks

At any time GovDAO by Supermajority Decision shall have the right and authority
to address bugs in the Core Software including the GnoVM, node software, and
systems Gno code by filtering or annotating transactions or making adjustments
to state (Transaction Replay Forks). This shall not apply to general Gno
ecosystem Gno code issues.

Before and until 1 year after $GNOT is transferrable GovDAO by Supermajority
Decision shall have the right and authority to make Transaction Replay Forks
for the purpose of addressing hacks, thefts, and exploits or other problems in
the spirit of the Constitution.

The extra-blockchain information of a Transaction Replay Fork such as
transaction annotations, filters, and state adjustments shall be considered a
part of the node software and not a part of the blockchain structure (of blocks
of transactions) itself, so as to preserve the Principle of the Chain.

These terms are expected to be modified before the 1 year mark to clarify or
modify the Transaction Replay Fork rules.

## Software Audits

GovDAO is responsible for ensuring that GnoWeb and Gno.land track audit reports
of all code uploaded to Gno as well as the code of the blockchain and VM
itself.

All auditing entity must be qualified; they must be vetted and whitelisted by
GovDAO Supermajority Decision on a Gno.land realm managed directly by GovDAO
(the Auditor and Audits Registry). The auditing entity must be vetted with
sufficient due dilligence and already have a good reputation and track record.
This is to prevent malicious audit reports from confusing and exploiting users.
The auditing reports by such Qualified Auditors are called Qualified Audit
Reports.

Each auditing entity must be registered with general information such as the
jusisdiction of the auditing entity and the public identity of its chief
executives; and significant owners (more than 10%) of the auditing entity must
be disclosed to GovDAO T1 members. The auditing entity may not be a DAO nor be
fully automated based on AI. This is to ensure that there is a clear legal
jurisdiction and accountability in case of abuse by any auditing entity.

The Auditor and Audits Registry must allow any sufficiently $GNOT bonded user
(Bonded Auditor) from submitting newly discovered vulnerabilities missed by any
audit report (whether qualified or not) and with some reasonable deposit per
Bonded Vulnerability Report to be used for assessment of the validity of the
submission. The submission should include a hash of the vulnerability report
but not the vulnerability itself. The procedure to determine these reports will
be defined by future Constitutional Amendments.

The Auditor and Audits Registry realm must track the scope of every audit
report and vulnerability report and record all known relevant information such
that any third party may independently derive a score for each auditing entity.

GnoWeb and all official alternatives (GnoWeb+Alts) must prominently link to any
relevant Qualified Audit Reports and hosted on the Auditor and Audits Registry
realm for all Gno hosted Gno code; but must not convey to the user any
warrantees.

Audit reports from non-qualified auditors may also be displayed on GnoWeb+Alts,
but they must be folded away and require at least one more click or user action
to show along with a prominent warning that those auditors have not been vetted
nor unqualified.

GnoWeb+Alts may display with any Gno code any overall audit status, badge, or
similar (Audit Status) derived from all Qualified Audits and any standing
Bonded Vulnerability Reports; however the Audit Status may not be numeric based
on any line-count or file-count or similar so as to be misleading to the user.
Code coverage of audits may be displayed as a percentage or numeric value(s)
but must be clearly separate from the Audit Status.

In no case may GnoWeb+Alts display the best Audit Status or a "green" or
"approved" status or badge or anything that can be construed as such for any
code unless the entire scope of the code and its dependencies have been
reviewed by at least one Qualified Audit Report and all Qualified Audit Reports
are free of any critical or major vulnerabilities or any issue that may result
in the loss of control or value by any user, and all Qualified Auditors are
still in good standing, and no Bonded Vulnerability Reports have been submitted
either that has not been dismissed explicitly as a false-positive for the code
and all of its dependencies. Furthermore such a status or badge or anything
that can be construed as such may only be displayed if manually reviewed by at
least three GovDAO members two of whom are T1 or T2 members; and they shall
each be responsible for ensuring that all the conditions are satisifed; however
they shall not be responsible for the correctness of each Qualified Audit
Report.

## Realm Upgrading

Realms must be clearly shown in GnoWeb to be upgradeable or not upgradeable
(immutable).

 * An immutable realm may not revert to being upgradeable.

 * No types declared in upgradeable realms may be persisted in immutable realms.

 * No upgradeable realms may be imported by immutable realms.

This creates a two-class system where immutable realms only depend on other
immutable realms. The state of immutable realms may be mutated by any logic as
allowed, but the upgrading of upgreadeable realms can not otherwise affect any
immutable realms.

Library packages declared in /p/ may not be upgraded.

Gnoweb must make clear to users of upgradeable realms that any funds depend on
the trust of the realm controller.

The only exception to the above rules may occur in response to a hack or
exploit. GovDAO by Supermajority Decision and pre-approval by the Oversight
Body may patch any realm package or library package if it is deemed to be in
the best interest of the Gno.land users and $GNOT economy.

TODO Rest to be determined after launch by a Constitutional Amendment.

## Name Registration

The primary purpose shall be for providing human readible names for package
paths (e.g. `org\_name` in `"gno.land/r/org_name/pkg_name"`).

Name registration will not intially be implemented.

There shall be at most one core name registration system for Gno.land, though
anyone may deploy their own for any purpose not managed by Gno.land, and as per
the Seven Mandates of Gno.land no name shall be required for any core services
(such as MsgAddPackage or MsgExec).

The name registration system shall satisfy the following:

 * Only lowercase alphabetic ASCII characters are allowed with underscores not
   as the first character.

 * The length is between 5 and 25 characters.

 * The name shall be canonicalized such that 'l' and 'i' are the same, and the
   canonicalized name shall be used to enforce uniqueness.

 * The canonicalized name shall not start with any of:
   * "gl" (and neither "gi", so as to be distinct from "g1..." addresses)
   * "gno"
   * "cosmos"
   * "atom"
   * "admin" (and neither "admln")
   * "support"
   * and more prefixes as determined by GovDAO.
   * and blacklists also determined by GovDAO.

 * Notwithstanding the above NT,LLC may reserve names that start with "gno".

 * The canonicalized name shall not include (as prefix or suffix or otherwise)
   the canonicalized form of any reserved names as defined by
   https://github.com/handshake-org/hsd/pull/819 without an on-chain
   transparent process approved by GovDAO to respect trademarks with priority
   given to US, then the EU; or only if the inclusion is deemed to be
   sufficiently unique by a reasonable algorithm implemented on chain but still
   preventing collisions in the canonicalized form; or on a case-by-case basis.

 * The registration of the name shall be permanent. Names may be disabled by a
   procedure defined by a Constitutional Amendment that defines the criteria
   for such disablement, but such names cannot be used again by anyone else.

 * The name registration takes learnings from the Handeshake protocol including
   https://github.com/handshake-org/hsd/issues/649

## Treasury diversification

Each of the above mentioned treasuries may be separately diversified by
supermajority vote of one proposal per treasury by GovDAO in accordance with
the Constitution--but all GovDAO members with any conflicts of tokens (except
$GNOT) above 3 months of salary for a typical senior software developer MUST
abstain from voting. If a quorum cannot possibly be reached due to conflict of
interests the Oversight Body may waive the quorum requirement.

Diversification for each treasury must adhere to an Treasury Diversification
Template listed explicitly in this Constitution by a future amendment that
declares the desired target ratio of Approved Treasury Diversification Tokens,
and the following:

 * Besides $GNOT or the top two dominant Bitcoin fork tokens (presently $BTC
   and $BCH but may change in the future) which shall always be approved, all
   other tokens of the Approved Treasury Diversifiction Tokens must be
   stablecoins FULLY backed by collateral.
   * Coins that are FULLY backed by silver via decentralized and publicly
     auditable depositories approved by GovDAO AND NewTendermint,LLC are also
     considered stablecoins.
   * Tether and all stablecoins backed by any amount of Tether are never
     permitted.
   * Tokens that lose their approved status must be replaced (sold) for
     approved tokens as soon as reasonable.
   * These subclauses and the parent clause cannot be removed by any
     constitutional amendment.

 * $PHOTON is not a stablecoin but it may be allowed by up to 20%.

 * $ATONE tokens or any other staking tokens are not allowed.

 * In no case shall the amount of $GNOT sold per treasury per month for
   diversification purposes exceed 1% of $GNOT of that treasury as measured at
   the beginning of each month.

 * In no case shall the total amount of $GNOT sold for all treasuries per month
   for diversification purposes exceed 50% of the average of ($GNOT inflation
   rate, and the past month's transaction fee revenue), with priority given to
   ValTreasury, then to PayTreasury, then to ServicesTreasury.

 * All trades for diversification purposes must be performed by audited and
   approved Gno logic for Gno AMM contracts running on Gno.land. No person or
   council may directly manage the diversification of treasury tokens except
   that GovDAO may delegate an oversight DAO to halt any diversification
   exchange transactions.

## Forking Gno.land

These are the procedures to create a Qualified Fork of Gno.land.

First, a fork proposal is voted on by all GovDAO members. The fork proposal
must be voted YES by more than 1/3 of the total voting power of GovDAO
(abstains do not count) by the end of the voting period of three weeks. There
is no commitment yet for voting YES on this initial fork proposal.

If such a fork proposal reaches more than 1/3 of voting power, a second fork
commitment proposal must be initiated with the same terms as the original
within a week after the first fork proposal's voting period has ended. Again
all GovDAO members vote on this second proposal. The second fork proposal must
also be voted YES by more than 1/3 of all the total voting power of GovDAO by
the end of the voting period of three weeks. Abstaining is equivalent to voting
NO.  Those who vote YES here are committing to join the others who voted YES
and split from those who voted NO or abstained if it also reaches 1/3 of voting
power--members may only serve on one chain.

If the second fork commitment proposal also reaches more than 1/3 of voting
power, two concurrent proposals must be proposed to determine which fork
retains the original Gno.land identity: one for the current chain, and another
for the proposed fork. Only T1 members may vote on these proposals; and they
may vote YES or NO on one or both proposals.

 * If the current chain wins a 2/3 supermajority of votes of the original T1
   members (but not the proposed fork), everyone who voted YES on the fork
   commitment proposal are ejected from the GovDAO membership set, and the
   proposed fork is called a Qualified Fork.

 * If the proposed fork wins a 2/3 supermajority of votes of the original T1
   members but not the current chain, everyone who did not vote YES on the fork
   commitment proposal are ejected from the GovDAO membership set to join
   their Qualified Fork, and the proposed fork retains the identity of the
   Gno.land chain.

 * If both proposals win a 2/3 supermajority, NewTendermint,LLC decides which
   gets to retain the identity of the chain (and the other is ejected from the
   GovDAO membership set to join their Qualified Fork).

 * If neither fork wins a 2/3 supermajority of votes, NewTendermint,LLC decides
   which fork gets to keep the identity of the chain, and whether the existing
   chain can retain the "Gno.land" name.
   * If neither can keep the "Gno.land" name, the "Gno.land" name is deprecated
     and cannot be used by any chain ever again; and the exiting chain may be
     required to change its chain ID; and the Gno.land domain will be managed
     by NewTendermint,LLC to point to both chains; and both chains shall be
     deemed to be Qualified Forks. Any royalty payments due to the original
     chain may be directed to either fork or split between the two forks as
     determined by NewTendermint,LLC; or even nullified but only if
     NewTendermint,LLC does not have any conflict of interest.

In all cases any new forks that include "Gno" in the name or chain ID must get
approved by NewTendermint,LLC.

All Qualified Forks may copy and use all of the state (including transaction
data) and code as from before the fork.

Non-qualified forks are not hereby probited from copying the state or code of
realms and packages of all forks of Gno.land. That is, everyone who submits
code or transactions to Gno.land or any of its forks are agreeing to allow (and
attesting to having the unencumbered rightsto allow) the code and resulting
blockchain state to be used freely as per the Gno Network GPL license (a
copyleft license fork of the AGPL3.0 but with modifications to allow for
"strong attribution"). However this is necessarily complicated when users
submit Gno code that is not owned by them, so no rights are explicitly granted
here by Gno.land or NewTendermint,LLC.

As with most of the software for the Gno.land stack including the GnoVM (except
for Tendermint2 which is Apache2.0) and derived works, all users of the Gno
code and blockchain state derived from Gno.land must abide by the same strong
attribution terms as for the Gno.land and GnoVM software at the time the code
is submitted--for example a non-qualified fork of Gno.land must give strong
attribution to all of its users to Gno.land or another Qualified Fork as
determined by the policy set forth by NewTendermint,LLC. NewTendermint,LLC may
determine that giving attribution to a Qualified Fork is sufficient (as opposed
to giving attribution to Gno.land), but a non-qualified fork will never benefit
from strong attribution in this way. The policies may be updated by
NewTendermint,LLC from time to time; and the goal is to finalize portions of
the policy such that they may be embedded into the Gno.land Constitution.

These terms shall be made clear in GnoWeb and gnocli for anyone submitting code
or any transaction. Specifically for anyone uploading realm or library packages
to Gno.land using gnocli they must first sign an approved CLA compatible with
the above and include a hash of the CLA with each transaction.

## Software Security

All software including Core Software funded for by Gno.land must be Fully
Audited before any release when funded by Gno.land. This condition and
guidelines for compliance must be made explicit in any contract or agreement
before any funding decisions are approved by any Core DAO.

Fully Audited means that all dependencies of the software are also audited by a
Qualified Auditor unless excempt by the Constitution or Law. This requirement
shall be relaxed for a period of 4 years after launch for existing software
unless otherwise stated by Law.

External dependencies shall be regulated such that every update to the external
dependency shall require audits as well or manual review of each minimal
security patch. No major or minor version upgrades shall be allowed
automatically for releases; that is, development branches may allow such
automated updates but the release process must include a procedure for vetting
any changes to dependencies since the last version.

Concurrently GovDAO must contribute to Atom.One such that the validators are
incentivized and over time required to run on their own dedicated machines. XXX
move.

### Software Finalization

All funding for Core Software must be toward the finalization of the software
for the major version number of that software.

 * Funding for Core Software must include an estimate for the future time of
   finalization of the software which may be any number of years into the
   future.

 * Software development that exceeds 7 years of development without
   finalization for the software for the major version number may be halted by
   the Oversight Body unless an extension of up to 2 years is granted by a
   Supermajority Decision of GovDAO. Extensions may be granted at any time, and
   may be to another development party.

 * After the finalization of the software the only funding allowed is for the
   maintenance of that software 

### Private Key Security

At no point shall a GovDAO or Gno.land fund for the development of any software
or team or anything else that allows or encourages the user to enter their
private key or mnemonic on any online computer or mobile network-capable
device. The only exception shall be for existing hardware wallets that allow
for bluetooth, as some of these devices already support bluetooth, as long as
bluetooth can be disabled.

Only 24 word mnemonics may be supported by Core Software; not 12 or 18.

Ephemeral private keys with limited capabilities and default reasonable
limitations on losses (in the case of theft or exploit) may be generated on
online computers or mobile network-capable devices, but no ephemeral keys may
be imported manually by the user nor converted into a mnemonic so as to prevent
any confusion. The term "mnemonic" shall only refer to their master private
key.

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
     disabling of network of bluetooth capabilities is not sufficient.

 * Encouragement to use a 52 deck of cards or 42 rolls of 20-sided dice or the
   equivalent rolls of 6-sided dice to generate custom entropy; with education
   that hardware wallets may be compromised to generate insecure private keys.

Gno.land nor GovDAO nor any entity receiving funding from Gno.land or GovDAO
may not sell any hardware devices except by approved retailers of the
manufactuer.

### Safety Wrapper Contracts

GovDAO shall ensure the timely development of conventions, protocols, and
libraries for realm and library logic to support the freezing of user accounts
or native tokens or application tokens/property in such a way to minimize harm
for unrelated parties.

For the purpose of protecting users from theft or loss resulting from exploits,
hacks, or even user error GovDAO shall fund for the development of multiple
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

### Quantum Proof Cryptography

As new post-quantum cryptography, the development of quantum computers, and
advancements in algorithms for breaking elliptic curve cryptography and hash
functions are ongoing; and it is generally agreed that hash functions are more
secure against quantum computers than elliptic curve cryptography, we will
first explore Merkle hash based commitments for user and validator accounts
that can be used to recover after "D-day"; and also prohibit user accounts from
the re-use of public keys (whether elliptic curve based or new quantum proof
algorithm based) so as to protect against future attacks by shielding the
public key from exposure via a hash function.

The following is the roadmap for quantum proofing Gno.land. It may be modified
with a Constitutional Amendment by GovDAO and also the pre-approval of the
Oversight Body.

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
     * Begin work with reputable cryptographers with recent relavant
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

No development or endorsement of quantum-proof modifications to Gno.land Core
Software or any consensus engine shall take place without the pre-approval of
NT,LLC (or its delegate) of its design specification. GovDAO must use a
reasonable amount of funds (under the typical salary for a software architect
for a year) to collect proposals for such specifications to be approved by
NT,LLC (or its delegate) before funding for its development.

### New Cryptographic Primitives or Implementations

Any new cryptographic primities or implementations must require NT,LLC
pre-approval unless NT,LLC designates this role to a Cryptography Committee
also approved by GovDAO with at least 3 T1 or T2 members who are respected and
have well-cited publications in the field of cryptography and at least a
doctorates degree in cryptography or relevant fields in mathematics, and at
least 1 T1 or T2 members who are equally expert in the field of quantum
computing.

Such new cryptographic primitives or implementations have a need that is not
met by the existing core software.

New cryptographic primitives for when an old primitive is deemeed
insufficiently secure (e.g. when quantum computers break secp256k1 given
pubkey) then both the old and new pubkeys shall be independently derived and
both required for a testing period of 7 years with significant economic
incentivecs such as with bounties rewarded for cracking challenge keys. This
helps prevent the adoption of primitives or implementations with unknown or
secret exploits from posing systemic risk.

### Formal Verification of Go/Gno.

Once the market cap of $GNOT exceeds $8B GovDAO must pass a Constitutional
Amendment to include provisions for the formal verification of Go or Gno
programs with the pre-approval of a reputable researcher with publications about
formal verification using open source tools already used by industry for formal
verification of programs.

If meaningful progress of such cannot be made during any 2 years of funding
another qualfiied researcher must be consulted for development with a new team.

## Hardware Security

### Development of Open Hardware

Once the market cap of $GNOT exceeds $5B GovDAO must pass a Constitutional
Amendment to include provisions for the development of 100% open hardware (and
100% open firmware/software) devices for the following categories in order:

 * A device that sits between an online computer or mobile network-capable
   device and hardware signer such that all communication between the
   aforementioned devices can be inspected (the Hardware Packet Inspector).
   This reduces the chances that a backdoor in the hardware wallet may result
   in theft or loss. If no suitable 100% open hardware CPU chip is available a
   Risk-V architecture based chip may be used instead until one is avaialble.

 * Hardware signer that is separate from but compatible the Hardware Packet
   Inspector. This may use a single chip that is proprietary to hold the
   private key and logic for unlocking based on pin, but it must be a module
   contained within completely open hardware and open firmware/source such that
   a vulnerability or backdoor of the single chip can be limited.

 * Validator hardware signer to preserve the private key of the validator and
   prevent double-signing.

 * CPU chip based on the Risk-V architecture (or related reputable open source
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

No one shall be required to use the these products.

### Forensic Analysis of Common Hardware

Furthermore once the market cap of $GNOT exceeds $5B GovDAO must pass a
Constitutional Amendment to include provisions for funding an in-house team to
accountabily verify the integrity of commonly used hardware products of
Gno.land users, especially the forensic analysis of the chips of those hardware
products. 

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
team shall be made public and free such that any team may use them to offer
their own forensic analysis.

# Appendix

## Common DAO Spec

This is based off of atomone-hub/genesis/CONSTITUTION.md and simplified.

These common DAO specifications shall apply for all Core DAOs unless otherwise
specified, Special Purpose DAOs, and all sub-DAOs of these DAOs. Other DAOs
that are not Core DAOs or Special Purpose DAOs or Descendants of these DAOs
need not implement these specifications.

## Common DAO Spec - version gno.land/genesis

All sub-DAOs have parent DAOs. The parent DAO of the Core DAOs and Special
Purpose DAOs are the governance of the AtomOne Hub itself. Therefore all Core
DAOs and Special Purpose DAOs as well as their sub-DAOs and the Hub Governance
itself altogether form a tree structure. The parent DAO and the parent DAO’s
parent DAO and so on, all the way up to Hub Governance are altogether called
the Ancestors of a DAO. The sub-DAOs and their sub-DAOs and so on are called
the Descendants of a DAO.

Every DAO, upon creation, must have a Charter (which is composed of Purpose and
Description), an initial set of Council members (which may be empty) and may
also have  Bylaws and Mandates, The Purpose and Description must be plaintext
files. The Bylaws and Mandates must be named plaintext files or folders of
plaintext files, or folders of folders.

A DAO’s Charter, Bylaws, and Mandates may be changed by a Simple Majority vote
from any of the DAO’s ancestors, except from AtomOne Hub Governance which shall
require a Supermajority vote.

A DAO’s Bylaws, Mandates, and the Bylaws and Mandates of its ancestor DAOs, the
relevant Laws, and this Constitution, altogether are called the Governing
Documents of the DAO. 

A DAO has a Council composed of zero or more members, with no maximum number of
members unless otherwise specified in its Governing Documents.

The Council of a DAO may change the Bylaws of the DAO, and otherwise make
Decisions on behalf of the DAO by passing Proposals.

A DAO may establish any number of sub-DAOs through the DAO Council’s Simple
Majority vote, with their own defined Charters and specific bylaws and
mandates, as necessity may arise and in accordance with the parent DAO Charter
and bylaws. Sub-DAOs are owned by and can be controlled by the parent DAO, and
members are also subject to the ancestor DAOs’ bylaws and mandates.

A Simple Majority in DAO governance is defined to be exactly “more than half”.
A Supermajority in DAO governance is defined to be exactly "two thirds or
more". This is distinct from a Supermajority in Hub Governance.

By default, unless specified otherwise in its Governing Documents, the
following rules shall apply for Council voting:

- each member shall have equal voting power (no member may occupy multiple
  seats)  
- a Council member may resign and thereby remove themselves from the Council  
- vote options are YES, NO, or ABSTAIN
- the tally denominator is the total number of voters (ABSTAINS do not count).
- voting for proposals are open until they are decided immediately by a
  supermajority of YES votes, or dismissed immediately by a majority of NO
  votes, or otherwise the voting period has passed.

By default, unless specified otherwise in its Governing Documents, the
following rules shall apply for Council membership election:

- the Council may elect one or more new members, and/or remove one or more
  members, by Super Majority vote. (self mutating).   
- the DAO’s Ancestors may modify the Council membership with a Super Majority
  vote.

Each DAO shall have an associated crypto address which can hold any number of
tokens.  

DAOs may operate with logic on core shards, or, represented as a m-of-n
multisig account on the AtomOne hub where the signers are each members of the
DAO’s council, where m is more than ½ n and also m is 3 or more. In all cases
financial transactions from the DAO’s treasury must follow the passage of
governance proposals on the DAO.

### Section 4.b: Core DAOs with Special Powers

As stated, the AtomOne Hub Governance structure will include several DAOs, to
provide guidance, oversight, and support for various aspects of governance and
operations. 

The Core DAO Council has the authority to make decisions for the Core DAO
through a Simple Majority vote of its members, as well as update the Bylaws,
but cannot update its Charter or Mandates. 

All Core DAOs and their sub-DAOs shall be composed of Cosmonauts, and the DAO
Councils be composed of Citizens. All Cosmonauts and Citizens of these DAOs
must have public and known real human identities.

Members are encouraged to engage in multiple DAOs within the AtomOne Hub to
foster collaboration, innovation, and the exchange of ideas. However, to
maintain the integrity of governance, members must declare their conflicts of
interest  and are required to comply with the conflict of interest policies
outlined in this Constitution, the Laws, and applicable Governing Documents.
Members must recuse themselves by not voting or voting ABSTAIN on proposals
that pose a conflict of interest.

The following are Core DAOs of AtomOne:

- Steering DAO and its subDAOs  
- Oversight DAOs and their subDAOs  
- Special Purpose DAOs and their subDAOs

While Core DAOs are non-person entities and may hold ATONE tokens in its
treasury, unlike other DAOs, Core DAOs may not stake ATONE tokens.

All Core DAOs must manage the financial resources of its treasury efficiently
and transparently at all times. 

#### Steering DAO

There is only ever one Steering DAO.

The Steering DAO is responsible for providing guidance and advisory annotations
on proposals and funding reports, helping to steer the overall direction and
priorities of the AtomOne Hub.

The Steering DAO can:

- publish publish periodic announcements on chain  
- annotate all core DAO proposals with advisory notes to provide context,
  guidance, and recommendations, offering a non-binding perspective to aid
  decision-making.  
- annotate funding reports to highlight key insights, concerns, and
  recommendations.  
- adjust AtomOne Hub Governance proposal timelines ("bend time") to provide
  additional time for consideration, though this power cannot be used to
  indefinitely prevent proposals from being addressed.  
- reduce the threshold needed to pass new laws from a Constitutional Majority
  to a Supermajority.


The mostly advisory nature of the Steering DAO's annotations ensures they guide
but do not dictate decisions, and specific limitations on the power to adjust
proposal timelines must be defined to prevent abuse.

Upon genesis there will be no Steering DAO. A Steering DAO may be created,
dismissed, or replaced by a Supermajority vote on AtomOne hub governance. 

For the purpose of cohesion, before the Steering DAO can be created (or
replaced), its initial (or new) set of Council members must have all agreed to
joining that Council by cryptographically signing a list of all the council
members, along with the DAO’s identifier.

#### Oversight DAOs

An Oversight DAO is responsible for ensuring that all laws, bylaws, mandates,
and core DAO governance proposals as well as AtomOne Hub governance proposals
within the scope of its mandate) are in compliance with their respective
Governing Documents. 

There can be one or more Oversight DAOs.

Any Oversight DAO can (within the scope of its mandate):

- veto any proposal or transaction of any Core DAO or AtomOne Hub with proof of
  its violation of Governing Documents.  
- adjust Core DAO or AtomOne Hub proposal timelines ("bend time") to provide
  additional time for consideration, ensure thorough review, and ensure all
  Core DAO and AtomOne Hub proposals comply with its Governing Documents.  
- as exceptions to the above rules, no Oversight DAO may affect any proposals
  to change any Oversight DAO’s Council membership, or any proposals to suspend
  or dismiss any Oversight DAO; both of which either require the passage of new
  law, or specific proposal types with the same criteria as passing new law.

The broad veto power of Oversight DAOs ensures that no single proposal can
override constitutional principles or violate the Constitution, but vetoes
should be used sparingly and with justification.

Upon genesis there will be no Oversight DAOs. An Oversight DAO may be created,
dismissed, or replaced by a Supermajority vote on AtomOne hub governance. 

